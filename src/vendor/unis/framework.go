package unis

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
    "net/http"
    "strings"
    "runtime"

	"github.com/golang/glog"
	"github.com/zieckey/goini"
	"github.com/valyala/fasthttp"
	"github.com/buaazp/fasthttprouter"

	"path/filepath"
	"unis/tcp"
	"unis/tcp/ldd"
)

// The default unis.instance
var DefaultFramework = &duxFramework
var duxFramework Framework

type Framework struct {
	Conf *goini.INI
	ConfigFilePath         string

	BufPond                 map[string]*sync.Pool // map[buffer_name]pool_pointer, pool's pool is pond.

	tcpAddr                 string            // The tcp server listen address
	udpAddr                 string            // The udp server listen address
	httpAddr                string            // The http server listen address

	tcpCodec                string            // The TCP codec type：ldd or t1v3l .
	tcpCodecFactory         tcp.TCPCodecFactory

    //the handler is called when the protocol(http,udp,tcp and so on) is quiting; main for watchSignal
	protocolCloseHandlers   map[string]func()
    proto_handler_rwlock    sync.RWMutex

	//the handler is called when the process is quiting; main for every module to release some resource gracefully
	moduleCloseHandlers     map[string]func()
    module_handler_rwlock   sync.RWMutex
	modules                 map[string]Module // map<module-name, Module>

	monitorStatusEnable     bool
	statusFilePath          string // The status.html file path

    max_concurrent_request  int

    client_ip_limit         string
    client_ipf              IPFilter

    debug                   bool
}

func init() {

	duxFramework.modules = make(map[string]Module)

	duxFramework.protocolCloseHandlers = make(map[string]func())
	duxFramework.moduleCloseHandlers = make(map[string]func())

    duxFramework.client_ip_limit = "0.0.0.0"
	duxFramework.debug = false
}

func (fw *Framework) RegisterModule(name string, m Module) error {

	if _, ok := fw.modules[name]; ok {
		return errors.New(name + " module arready exists!")
	}

	fw.modules[name] = m
	return nil
}

func (fw *Framework) NewBufPool(poolName string, newObj func() interface{}) (*sync.Pool, error) {

	if pool, ok := fw.BufPond[poolName]; ok {
		return pool, errors.New(poolName + " have been exist.")
	}
	pool := &sync.Pool{New: newObj}
	fw.BufPond[poolName] = pool
	return pool, nil
}

func (fw *Framework) Initialize() error {

	if !flag.Parsed() {
		flag.Parse()
	}

	////////////// loading app.conf config ///////////////////////////
    fmt.Printf("loading config file of ckproxy is %s\n", *ConfPath)
    configFilePath := *ConfPath
    if _, err := os.Stat(*ConfPath); os.IsNotExist(err) {
        fmt.Printf("config file of ckproxy %s is not exist\n", *ConfPath)
        return err
    }

	fw.ConfigFilePath = configFilePath
	ini, err := goini.LoadInheritedINI(configFilePath)
	if err != nil {
		return errors.New("parse INI config file error : " + configFilePath)
	}

	fw.Conf = ini

	if v, ok := fw.Conf.SectionGetBool("common", "debug"); ok {
		fw.debug = v
	}

    if fw.debug  {
        flag.Set("log_dir", "logs")
        flag.Set("alsologtostderr", "true")

        go func() {
            glog.Info(http.ListenAndServe(":7090", nil))
        }()
    }
    flag.Set("v", "3")

	fw.BufPond = make(map[string]*sync.Pool)

	udpPort, _ := fw.Conf.SectionGet("common", "udp_port")
	tcpPort, _ := fw.Conf.SectionGet("common", "tcp_port")
	httpPort, _ := fw.Conf.SectionGet("common", "http_port")

	if len(udpPort) == 0 && len(tcpPort) == 0 && len(httpPort) == 0 {
		return errors.New("not found communication port")
	}

	if len(httpPort) > 0 {
		fw.httpAddr = fmt.Sprintf(":%v", httpPort)
	}

	if len(tcpPort) > 0 {
		fw.tcpAddr = fmt.Sprintf(":%v", tcpPort)
	}

	if len(udpPort) > 0 {
		fw.udpAddr = fmt.Sprintf(":%v", udpPort)
	}

	fw.max_concurrent_request, _ = fw.Conf.SectionGetInt("common", "max_concurrent_request")
    glog.Infof("max_concurrent_request=%d", fw.max_concurrent_request )

	var ok bool
	fw.monitorStatusEnable, ok = fw.Conf.SectionGetBool("common", "monitor_status_enable")
	if !ok {
		fw.monitorStatusEnable = true
	}
	fw.statusFilePath = fw.GetPathConfig("common", "monitor_status_file_path")

	if err = fw.initTCPCodec(); err != nil {
		return err
	}

    /////////////////add ip limit for client ///////////////////////////////
    ips_limit, _ := fw.Conf.SectionGet("common", "client_ip_limit")
	if len(ips_limit) == 0 {
        ips_limit = "0.0.0.0"
    }
    fw.client_ip_limit = ips_limit
    glog.Infof("client_ip_limit is %s", ips_limit)

    var data = []byte(ips_limit)
    if err := fw.client_ipf.Load( data ); err != nil {
        glog.Infof("load client ip filter list failed %v", err)
        return err
    }

    local_addrs, err := net.InterfaceAddrs()
    if err != nil {
        os.Stderr.WriteString("Oops:" + err.Error())
        os.Exit(1)
    }
    for _, la := range local_addrs {
        if ipnet, ok := la.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                fw.client_ipf.AddIPString( ipnet.IP.String() )
                glog.Infof("add local internal addr %s to client_ip_limit", ipnet.IP.String() )
            }
        }
    }
    ///////////////////////////////////////////////////////////////////////

    return nil
}

func (fw *Framework) runHTTP() {

	defer wg.Done()

	glog.Infof("Running http service at %s\n", fw.httpAddr)

	listener, err := net.Listen("tcp", fw.httpAddr)
	if err != nil {
		glog.Errorf("%s", err.Error())
	}

	runloop := true
    fw.proto_handler_rwlock.Lock()
	fw.protocolCloseHandlers["http"] = func() {
		l := listener
		l.Close()
		runloop = false
	}
    fw.proto_handler_rwlock.Unlock()

	max_http_concurrent_connections, _ := fw.Conf.SectionGetInt("common", "max_http_concurrent_connections")
	max_http_concurrent_conn_per_ip, _ := fw.Conf.SectionGetInt("common", "max_http_concurrent_conn_per_ip")

	svr := &fasthttp.Server{
		Handler        : fasthttp_router.Handler,
		Name           : "skylar cloudkill http proxy",
		Concurrency    : max_http_concurrent_connections,
    	MaxConnsPerIP  : max_http_concurrent_conn_per_ip,
	}

	if err := svr.Serve(listener); err != nil {
    	glog.Fatalf("error in Serve: %s", err.Error())
	}

}

func (fw *Framework) runTCP() {

	defer wg.Done()
	glog.Infof("Running tcp service at %s\n", fw.tcpAddr)

	if len(fw.tcpAddr) == 0 {
		return
	}

	a, err := net.ResolveTCPAddr("tcp", fw.tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", a)
	if err != nil {
		log.Fatal(err)
	}

	runloop := true
    fw.proto_handler_rwlock.Lock()
	fw.protocolCloseHandlers["tcp"] = func() {
		l := listener
		l.Close()
		runloop = false
	}
    fw.proto_handler_rwlock.Unlock()

    var lastOverflowErrorTime time.Time
    var wp *WorkerPool

    enable_worker_pool, _ := fw.Conf.SectionGetBool("common", "enable_worker_pool")
    enable_worker_pool = false
    if enable_worker_pool {
	    max_worker_count, _ := fw.Conf.SectionGetInt("common", "max_worker_count")
	    max_worker_idle_time, _ := fw.Conf.SectionGetInt("common", "max_worker_idle_time")

        glog.Infof("enable worker pool=%v for TCP Channel", enable_worker_pool)
        glog.Infof("\tmax_worker_count=%v", max_worker_count)
        glog.Infof("\tmax_worker_idle_time=%v", max_worker_idle_time)
	    wp, _ = NewWorkerPoolLimit(max_worker_count, max_worker_idle_time)
    }

	for {
		c, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			if runloop == false {
				log.Println("tcp quit")
				return
			}
			continue
		}

        is_ok := fw.check_client_ipf( c.RemoteAddr().String() )
        if !is_ok {
		    glog.Infof(" remote host %v is not allow to access ckproxy", c.RemoteAddr().String())
			c.Close()
            return
        }

        if fw.debug {
		    glog.Infof("new connection come from %v", c.RemoteAddr().String())
        }

        if enable_worker_pool {

		    f := func() {
                if fw.debug {
			        glog.Infof("starting new worker for remoteAddr %v", c.RemoteAddr().String())
                }
			    serveTCP(c)
            }

            if !wp.Serve(f) {

                glog.Infof("The connection cannot be served because Server.Concurrency limit exceeded")

                c.Close()

                if time.Since(lastOverflowErrorTime) > time.Minute {
                    glog.Infof("The incoming connection cannot be served, because %d concurrent connections are served. "+
                    "Try increasing Server.Concurrency", wp.max_worker_count)
                    lastOverflowErrorTime = time.Now()
                }

                // The current server reached concurrency limit,
                // so give other concurrently running servers a chance
                // accepting incoming connections on the same address.
                //
                // There is a hope other servers didn't reach their
                // concurrency limits yet :)
                time.Sleep(100 * time.Millisecond)
            }

         } else {

	        if runtime.NumGoroutine() < DefaultFramework.max_concurrent_request {
		        go serveTCP(c)
	        } else {
                if fw.debug {
                    glog.Infof("The connection cannot be served because Server.Concurrency limit exceeded")
                }
		        runtime.GC()
		        c.Close()
	        }

        }

        c = nil
	}
}

func (fw *Framework) runUDP() {

	defer wg.Done()

	glog.Infof("Running udp service at %s\n", fw.udpAddr)

	if len(fw.udpAddr) == 0 {
		return
	}

	pool, err := fw.NewBufPool("UDP", func() interface{} { return make([]byte, 1472) })
	if err != nil {
		glog.Errorf("%s", err.Error())
	}

	a, err := net.ResolveUDPAddr("udp", fw.udpAddr)
	if err != nil {
		glog.Errorf("%s", err.Error())
	}

	listener, err := net.ListenUDP("udp", a)
	if err != nil {
		glog.Errorf("%s", err.Error())
	}

	runloop := true
    fw.proto_handler_rwlock.Lock()
	fw.protocolCloseHandlers["udp"] = func() {
		l := listener
		l.Close()
		runloop = false
	}
    fw.proto_handler_rwlock.Unlock()

    var lastOverflowErrorTime time.Time
    var wp *WorkerPool
    enable_worker_pool, _ := fw.Conf.SectionGetBool("common", "enable_worker_pool")
    if enable_worker_pool {
	    max_worker_count, _ := fw.Conf.SectionGetInt("common", "max_worker_count")
	    max_worker_idle_time, _ := fw.Conf.SectionGetInt("common", "max_worker_idle_time")

        glog.Infof("enable worker pool=%v for UDP Channel", enable_worker_pool)
        glog.Infof("\tmax_worker_count=%v", max_worker_count)
        glog.Infof("\tmax_worker_idle_time=%v", max_worker_idle_time)

	    wp, _ = NewWorkerPoolLimit(max_worker_count, max_worker_idle_time)
    }

	for {

		buf := pool.Get().([]byte)
		rlen, remoteAddr, err := listener.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			if !runloop {
				log.Println("udp quit")
				return
			}
			continue
		}

        is_ok := fw.check_client_ipf( remoteAddr.String() )
        if !is_ok {
            return
        }

        if fw.debug {
            glog.Infof("packet come from %v, %s, len %v", remoteAddr, buf[:rlen], rlen)
        }

        if enable_worker_pool {

		    f := func() {
                if fw.debug {
			        glog.Infof("starting new worker for remoteAddr %v", remoteAddr)
                }
			    serveUDP(listener, remoteAddr, buf, rlen)
            }

            if !wp.Serve(f) {

                glog.Infof("The UDP request from client cannot be served because Server.Concurrency limit exceeded")
                if time.Since(lastOverflowErrorTime) > time.Minute {
                    glog.Infof("The incoming UDP request from client cannot be served, because %d concurrent requests are served. "+
                    "Try increasing Server.Concurrency", wp.max_worker_count)
                    lastOverflowErrorTime = time.Now()
                }

                // The current server reached concurrency limit,
                // so give other concurrently running servers a chance
                // accepting incoming request on the same address.
                // There is a hope other servers didn't reach their
                // concurrency limits yet :)
                time.Sleep(100 * time.Millisecond)

            }

        } else {

		    if runtime.NumGoroutine() < DefaultFramework.max_concurrent_request {
			    go serveUDP(listener, remoteAddr, buf, rlen)
		    } else {
                if fw.debug {
                    glog.Infof("The UDP request from client cannot be served because Server.Concurrency limit exceeded")
                }
                runtime.GC()
			    continue
		    }
        }
	}
}

func (fw *Framework) watchSignal() {

    defer wg.Done()

	c := make(chan os.Signal, 1)
	signal.Notify(c)

	go func() {
		defer close(c)
		for {
			s := <-c
			glog.Errorf("Got signal %v", s)
			if s == syscall.SIGHUP || s == syscall.SIGINT || s == syscall.SIGTERM {

                fw.proto_handler_rwlock.RLock()
				for _, functor := range fw.protocolCloseHandlers {
					functor()
				}
                fw.proto_handler_rwlock.RUnlock()

                fw.module_handler_rwlock.RLock()
				for _, functor := range fw.moduleCloseHandlers {
					functor()
				}
                fw.module_handler_rwlock.RUnlock()

			}
		}
	}()
}

var (
	fasthttp_router *fasthttprouter.Router
	wg sync.WaitGroup
)

func (fw *Framework) Run() {

	fw.createPidFile()
	defer fw.removePidFile()

	//////////// register internal module //////////////////////////////
	if fw.monitorStatusEnable {
		fw.RegisterModule("monitor", new(MonitorModule))
	}
	////////////////////////////////////////////////////////////////////
	fasthttp_router = fasthttprouter.New()
	///////////////////register business module/////////////////////////
	for name, module := range fw.modules {
		err := module.Initialize()
		if err != nil {
			log := name + " module initialized failed : " + err.Error()
			glog.Errorf("%v", log)
			panic(log)
		}
	}
	////////////////////////////////////////////////////////////////////

	wg.Add(1)
	fw.watchSignal()

	wg.Add(1)
	go fw.runHTTP()

	wg.Add(1)
	go fw.runUDP()

    /***
    wg.Add(1)
	go fw.runTCP()
    ***/

	wg.Wait()

	////////////////////////////////////////////////////////////////////
}

/**
 * GetPathConfig 获取一个路径配置项的相对路径（相对于 ConfPath 而言）
 * e.g. :
 * 		ConfPath = /home/unis.conf/app.conf
 *
 *	and the app.conf has a config item as below :
 *  	[business]
 *		qlog_conf = qlog.conf
 *
 * and then the GetPathConfig("business", "qlog_conf") will
 * return /home/unis.conf/qlog.conf
 */

func (fw *Framework) GetPathConfig(section, key string) string {

	filepath, ok := fw.Conf.SectionGet(section, key)
	if !ok {
		println(key + " config is missing in " + section)
		return ""
	}
	return goini.GetPathByRelativePath(fw.ConfigFilePath, filepath)
}

const (
	LddTcpCodec  = "ldd"
	Tcp1V3LCodec = "t1v3l" // 1 byte version and 3 bytes length. The protocol used by VPN project
)

func (fw *Framework) initTCPCodec() error {

	fw.tcpCodec, _ = fw.Conf.SectionGet("common", "tcp_codec")
	switch fw.tcpCodec {
	case LddTcpCodec:
		fw.tcpCodecFactory = ldd.NewLddTCP
		break
	case Tcp1V3LCodec:
		// TODO
		break
	}
	return nil
}

func (fw *Framework) createPidFile() {

	pidpath := fw.GetPathConfig("common", "pid_file")
	path := filepath.Dir(pidpath)

	if err := os.MkdirAll(path, 0777); err != nil {
	}
	pid := os.Getpid()
	pidString := strconv.Itoa(pid)

	if err := ioutil.WriteFile(pidpath, []byte(pidString), 0777); err != nil {
		panic("Create pid file failed : " + pidpath)
	}
}

func (fw *Framework) removePidFile() {

	pidpath := fw.GetPathConfig("common", "pid_file")
	os.Remove(pidpath)
	println("remove pid file : ", pidpath)

}

func (fw *Framework) RegisterModuleCloseHandler(functor func(), desc string) {
	fw.moduleCloseHandlers[desc] = functor
}

func (fw *Framework) check_client_ipf( remoteAddr string ) bool {

    if strings.Compare(fw.client_ip_limit, "0.0.0.0") == 0 {
        return true
    }

    pos := strings.LastIndex(remoteAddr, ":")
    remote_peer_ip := remoteAddr[:pos]
    if !fw.client_ipf.FilterIPString( remote_peer_ip ) {
        if fw.debug {
            glog.Infof("remote client %s is not allow to access ckproxy\n", remoteAddr)
        }
        return false
    }
    return true
}

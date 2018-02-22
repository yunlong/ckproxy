package ckproxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
    	"unis"

	"github.com/golang/glog"
//	"golib/cgo/qhsec"
)

const (
	HTTP = iota
	UDP
	TCP
)

type CKProxy struct {
    httpClient     http.Client

	RecoverySec    int
	OfflineFailedCount int

	UdpTimeoutMs       int
	HttpTimeoutMs      int

	HttpRetryTimes int
	UdpRetryTimes  int

	httpProxy map[string]RouterInterface
	udpProxy  map[string]RouterInterface

	//nppConfig *qhsec.NppConfig
}

type LddContext struct {
	fctx  *unis.Context
	proto int
}

var defaultCKProxy *CKProxy

func Instance() (ck *CKProxy) {
	return defaultCKProxy
}

func (ck *CKProxy) Initialize() error {
	defaultCKProxy = ck
	fw := unis.DefaultFramework

	logLevel, _ := fw.Conf.SectionGet("common", "log_level")
	var l glog.Level
	l.Set(logLevel)

	ck.RecoverySec, _ = fw.Conf.SectionGetInt("common", "recovery_sec")
	ck.OfflineFailedCount, _ = fw.Conf.SectionGetInt("common", "offline_failed_count")
	ck.HttpTimeoutMs, _ = fw.Conf.SectionGetInt("common", "http_timeout_ms")
	ck.UdpTimeoutMs, _ = fw.Conf.SectionGetInt("common", "udp_timeout_ms")

	if ck.RecoverySec == 0 {
		ck.RecoverySec = 1800
	}
	if ck.OfflineFailedCount == 0 {
		ck.OfflineFailedCount = 10
	}
	if ck.UdpTimeoutMs == 0 {
		ck.UdpTimeoutMs = 1000
	}

	///////////////////////////////////////////////////////////////////////////
	ck.HttpRetryTimes, _ = fw.Conf.SectionGetInt("common", "http_retry_times")
	ck.UdpRetryTimes, _ = fw.Conf.SectionGetInt("common", "udp_retry_times")
	///////////////////////////////////////////////////////////////////////////

	glog.Infof("==================== skylar cloud kill proxy setting ===================")

	glog.Infof("recovery_sec=%v", ck.RecoverySec)
	glog.Infof("offline_failed_count=%v", ck.OfflineFailedCount)
	glog.Infof("http_timeout_ms=%v", ck.HttpTimeoutMs)
	glog.Infof("udp_timeout_ms=%v", ck.UdpTimeoutMs)
	glog.Infof("recovery_sec=%v", ck.RecoverySec)
	glog.Infof("http_retry_times=%v", ck.HttpRetryTimes)
	glog.Infof("udp_retry_times=%v", ck.UdpRetryTimes)

    glog.Infof("==================== http proxy router interface =======================")
    glog.Infof("http proxy router interface:")
	ck.httpProxy = make(map[string]RouterInterface)
	proxy, _ := fw.Conf.GetKvmap("http_proxy")
	for k, v := range proxy {
		r := new(Router)
		r.Url = v
		r.Type = HTTP
		r.Status = true
		ck.httpProxy[k] = r
		unis.HandleFunc(k, unis.DefaultPlain, ck.HttpProcess, ck)
	//	unis.HandleFunc(k, unis.DefaultCipher, ck.HttpProcess, ck)
		glog.Infof("%s=%s\n", k, r.Url)
	}

    /**
     * udp netmethod
     * 0：cloudquery(openapi也在这个接口中)
     * 1: getconf(也会调用新qconf)
     * 2: stats，fixed-get(查几个特殊md5的值)
     * 3: status (监控，qdns调用，判断服务是否存活)
     * 4: echo服务(将请求的密文数据解包后，以明文返回请求内容)
     * 6：新版qconf
     * 11：专用打点服务
     */

    glog.Infof("==================== udp proxy router interface =======================")
	ck.udpProxy = make(map[string]RouterInterface)
	proxy, _ = fw.Conf.GetKvmap("udp_proxy")
	for k, v := range proxy {
		r := new(Router)
		r.Url = v
		r.Type = UDP
		r.Status = true
		ck.udpProxy[k] = r

		unis.UDPHandleFunc(k, unis.DefaultPlain, ck.UdpProcess, ck)
		glog.Infof("%s=%s\n", k, r.Url)
	}
	glog.Infof("=======================================================================")

	unis.HandleFunc("/ckproxy/status", unis.DefaultPlain, ck.Stat, ck)

	unis.HandleFunc("/api/system_repair.json", unis.DefaultPlain, ck.SystemRepairProcess, ck)
	{
		r := new(Router)
		r.Url = "http://qup.f.360.cn/api/system_repair.json"
		r.Type = HTTP
		r.Status = true
		ck.httpProxy["/api/system_repair.json"] = r
	}

	err := ck.InitNppConfig()
    if err != nil {
        return err
    }

	return nil
}
/**
func (ck *CKProxy) NppConfig() *qhsec.NppConfig {
    return ck.nppConfig
}
**/

func (ck *CKProxy) InitNppConfig() error {

    /**
    fw := unis.DefaultFramework
    business_name, _ := fw.Conf.SectionGet("safe_proxy", "business_name")
    symmetric_keys_path := fw.GetPathConfig("safe_proxy", "symmetric_keys_path")
    asymmetric_keys_path := fw.GetPathConfig("safe_proxy", "asymmetric_keys_path")

	glog.Infof("loading symmtric keys %s", symmetric_keys_path)
	glog.Infof("loading asymmtric keys %s", asymmetric_keys_path)

    p, err := qhsec.NewNppConfig(symmetric_keys_path, business_name, asymmetric_keys_path)
    if err != nil {
        return err
    }
    ck.nppConfig = p
    **/

    return nil
}

func (ck *CKProxy) HttpProcess(recv []byte, ctx *unis.Context) (send []byte, err error) {

	router, ok := ck.httpProxy[ctx.Method]
	if !ok {
		return nil, fmt.Errorf("the proxy [%v] not found", ctx.Method)
	}

	glog.Infof("%v netmethod get url %v", ctx.RemoteAddr, router.NextAddr())

	if !router.IsOnline() {
		glog.Errorf("%v offline", router.String())
		return nil, fmt.Errorf("the proxy [%v] offline", ctx.Method)
	}

	body := bytes.NewReader(recv)
	req, err := http.NewRequest(string(ctx.HttpReq.Method()), router.NextAddr(), body)
    if err != nil {
		router.AddFaileCount(1)
		return nil, err
	}

	ctxhdr := make(http.Header)
	ctx.HttpReq.Request.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case "Transfer-Encoding":
			req.TransferEncoding = append(req.TransferEncoding, sv)
		default:
			ctxhdr.Set(sk, sv)
		}
	})
	req.Header = ctxhdr

	var resp *http.Response = nil
	for retry := 0; retry < ck.HttpRetryTimes; retry++ {
		resp, err = ck.httpClient.Do(req)
		if err == nil {
			break
		} else {
			router.AddFaileCount(1)
		}
	}

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		router.AddSuccessCount(1)
		return ioutil.ReadAll(resp.Body)
	}

	router.AddFaileCount(1)
	if resp.StatusCode == 404 { // 403 can be rate limit error.  || resp.StatusCode == 403 {
		err = fmt.Errorf("remote 404, resource not found: %s", router.NextAddr())
	} else {
		err = fmt.Errorf("remote %s -> %d", router.NextAddr(), resp.StatusCode)
	}

	return send, err
}

func (ck *CKProxy) UdpProcess(recv []byte, ctx *unis.Context) (send []byte, err error) {

	router, ok := ck.udpProxy[ctx.Method]
	if !ok {
		return nil, fmt.Errorf("the proxy [%v] not found", ctx.Method)
	}

	glog.Infof("%v netmethod get url %v", ctx.RemoteAddr, router.NextAddr())
	conn, err := net.Dial("udp", router.NextAddr())
	timeout := time.Now().Add(time.Duration(ck.UdpTimeoutMs) * time.Millisecond)
	conn.SetDeadline(timeout)
	defer conn.Close()
	if err != nil {
		router.AddFaileCount(1)
		glog.Errorf("%v get error %v", conn.RemoteAddr().String(), err)
		return nil, err
	}

	ret, err := conn.Write(recv)
	if err != nil {
		router.AddFaileCount(1)
		glog.Errorf("%v write error %v", conn.RemoteAddr().String(), err)
		return nil, err
	}

	var buf [4096]byte
	ret, err = conn.Read(buf[0:])
	if err != nil {
		router.AddFaileCount(1)
		glog.Errorf("%v read error %v", conn.RemoteAddr().String(), err)
		return nil, err
	}

	send = buf[0:ret]
	router.AddSuccessCount(1)

	return send, err
}

func (ck *CKProxy) Stat(recv []byte, ctx *unis.Context) ([]byte, error) {

	str := string("http:\n")
	for _, r := range ck.httpProxy {
		str += r.String()
		str += "\n"
	}
	str += string("udp:\n")
	for k, r := range ck.udpProxy {
		str += k + "=" + r.String()
		str += "\n"
	}

	return []byte(str), nil
}

func (ck *CKProxy) SystemRepairProcess(recv []byte, ctx *unis.Context) (send []byte, err error) {

	router, ok := ck.httpProxy[ctx.Method]
	if !ok {
		return nil, fmt.Errorf("the proxy [%v] not found", ctx.Method)
	}

    glog.Infof("%v netmethod get url %v", ctx.RemoteAddr, router.NextAddr())

	if !router.IsOnline() {
		glog.Errorf("%v offline", router.String())
		return nil, fmt.Errorf("the proxy [%v] offline", ctx.Method)
	}

	var url string
	if string(ctx.HttpReq.Method()) == "GET" {
        // for get request format 1
		val := ctx.HttpReq.URI().QueryArgs().Peek("url")
		if val != nil {
			url = string(val)
			glog.Info("get url :%v", url)
		} else {
			glog.Errorf("%v offline", router.String())
			return nil, fmt.Errorf("request format error, %v %v", ctx.Method, ctx.HttpReq.URI())
		}
	} else {
        // post request format 3
		val := ctx.HttpReq.URI().QueryArgs().Peek("url");
        if val != nil {

			url = string(val)
			glog.Info("post parameter get url :%v", url)

		} else {

            body := bytes.NewReader(recv)
	        req, err := http.NewRequest(string(ctx.HttpReq.Method()), url, body)
	        if err != nil {
				return nil, fmt.Errorf("post parse request format error, %v %v %v ", ctx.Method, ctx.HttpReq.URI(), err)
	        }

            ctxhdr := make(http.Header)
            ctx.HttpReq.Request.Header.VisitAll(func(k, v []byte) {
                sk := string(k)
                sv := string(v)
                switch sk {
                case "Transfer-Encoding":
                    req.TransferEncoding = append(req.TransferEncoding, sv)
                default:
                    ctxhdr.Set(sk, sv)
                }
            })
            req.Header = ctxhdr

			val := req.PostFormValue("url")
			//fmt.Printf("val %v\n", ctx.HttpReq.Form)
			if len(val) == 0 {
				return nil, fmt.Errorf("request format error, parameter url is empty, %v %v %v ", ctx.Method, ctx.HttpReq, err)
			}
			url = val
			glog.Info("post multipart parameter get url :%v ", url)
		}
	}

	body := bytes.NewReader(recv)
	req, err := http.NewRequest(string(ctx.HttpReq.Method()), url, body)
	if err != nil {
		router.AddFaileCount(1)
		return nil, err
	}

	ctxhdr := make(http.Header)
	ctx.HttpReq.Request.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case "Transfer-Encoding":
			req.TransferEncoding = append(req.TransferEncoding, sv)
		default:
			ctxhdr.Set(sk, sv)
		}
	})
	req.Header = ctxhdr

    var resp *http.Response = nil
	for retry := 0; retry < ck.HttpRetryTimes; retry++ {
		resp, err = ck.httpClient.Do(req)
		if err == nil {
			break
		} else {
			router.AddFaileCount(1)
		}
	}

	if err != nil {
		router.AddFaileCount(1)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		router.AddSuccessCount(1)
		return ioutil.ReadAll(resp.Body)
	}

	router.AddFaileCount(1)
	if resp.StatusCode == 404 { // 403 can be rate limit error.  || resp.StatusCode == 403 {
		err = fmt.Errorf("remote 404, resource not found: %s", router.NextAddr())
	} else {
		err = fmt.Errorf("remote %s -> %d", router.NextAddr(), resp.StatusCode)
	}
	return send, err
}

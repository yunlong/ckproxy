package unis

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
)

var tcpHandlers = make(map[int]*Handler)

func handleTcp(netMethod int, h *Handler) {
	tcpHandlers[netMethod] = h

	// 网络接口号 0 作为特殊用途，如下：
	// 针对 LDD 长连接协议中ping/pong类型的请求的特殊处理。
	if duxFramework.tcpCodec == LddTcpCodec {
		tcpHandlers[0] = h
	}
}

func handleOnConn(netMethod int, onNew OnConnFunc, onClose OnConnFunc) {
	h := tcpHandlers[netMethod]
	h.onNewConn = onNew
	h.onCloseConn = onClose
}

// 为了在框架中处理LDD长连接协议中的ping/pong类型的请求，实现一个闭包形式的callback，避免了循环依赖。
// 框架调用 TCPCodec.Handle 进而会调用LDD的具体实现函数 LddTCP.Handle，在该函数中，再回调当前的Handler(应用层注册)，从而统一处理tcp数据包
type lddTCPHandler struct {
	c *Context
	h *Handler
}

func (t *lddTCPHandler) CallbackHandle() (send []byte, err error) {
	d := t.c.TCPAccessor.Payload()

	/***
	var unpacker *qhsec.ServerUnpacker
	if t.h.cipherFlag&DecryptRequest > 0 {
		unpacker, err = qhsec.NewServerUnpacker(t.h.m.NppConfig())
		if err != nil {
			glog.Errorf("%v qhsec.NewServerUnpacker error : %v", t.c.Conn.RemoteAddr(), err.Error())
			return nil, err
		}
		dec, err := unpacker.Unpack(d)
		if err != nil {
			glog.Errorf("qhsec.NewServerUnpacker decrypt message from %v error : %v", t.c.Conn.RemoteAddr(), err.Error())
			return nil, err
		}
		d = dec

		if duxFramework.debug {
			glog.Infof("http body decrypted : %v", base64.StdEncoding.EncodeToString(d))
		}
	}
	***/

	send, err = t.h.f(d, t.c)
	if err != nil {
		return send, err
	}

	d = send

	/***
	if t.h.cipherFlag&EncryptResponse > 0 {
		d, err = unpacker.Pack(d)
		if err != nil {
			glog.Errorf("%v application qhsec.Pack error : %v", t.c.Conn.RemoteAddr(), err.Error())
			return send, err
		}
	}
	***/

	return d, nil
}

func serveTCP(lconn *net.TCPConn) {

	defer lconn.Close()
	var lock sync.Mutex

	var onNew OnConnFunc
	var onClose OnConnFunc

	hasReportConn := false

//	stat := stat.NewStatHelper()
//	defer stat.Close()

	for {

		if duxFramework.tcpCodecFactory == nil {
			glog.Warningf("tcpCodecFactory nil ERROR ")
			return
		}

		codec := duxFramework.tcpCodecFactory()
		//codec = ldd.NewLddTCP() // TODO use another Factory function to create TCPCodec
		err := codec.ReadHeader(lconn)
		if err != nil {
			if err == io.EOF {
				glog.Warningf("Connection from %v closed.", lconn.RemoteAddr())
			} else {
				glog.Warningf("ReadHeader ERROR : %v", err.Error())
			}
			break
		}
		err = codec.ReadBody(lconn)
		if err != nil {
			glog.Warningf("ReadBody ERROR : %v", err.Error())
			break
		}

		netMethod := int(codec.PayloadType())
		h, ok := tcpHandlers[netMethod]
		if !ok {
			glog.Warningf("cannot find HANDLER to process this message. PayloadType=%v", codec.PayloadType())
			continue
		}

		if !hasReportConn {
			onNew = h.onNewConn
			onClose = h.onCloseConn

			if onNew != nil {
				(onNew)(lconn)
				hasReportConn = true
			}
		}

		ctx := new(Context)
		ctx.Method = strconv.Itoa(netMethod)
		ctx.Proto = TCP
		ctx.Conn = lconn
		ctx.HttpReq = nil
		ctx.TCPAccessor = codec
		ctx.RemoteAddr = lconn.RemoteAddr().String()

		th := new(lddTCPHandler)
		th.c = ctx
		th.h = h

		go func() {

			beginTime := time.Now()
			send, err := codec.Handle(th)
			if err != nil {
				glog.Warningf("handler ERROR : %v", err.Error())
				//continue
				return
			}

			lock.Lock()
			err = codec.Write(send, lconn)
			lock.Unlock()

			if err != nil {
				fmt.Printf("write err %v", err)
			}

			if duxFramework.debug && ctx.Method != "0" {
			    costMs := float64(time.Since(beginTime).Nanoseconds()) / 1000000.0
				glog.Infof("tcp method:%v\t,remote:%v\tcost:%v ms", ctx.Method, ctx.Conn.RemoteAddr(), costMs)
			}

		}()
	}

	if onClose != nil {
		(onClose)(lconn)
	}
}

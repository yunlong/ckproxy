package unis

import (
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
)

var udpHandlers = make(map[int]*Handler)

func handleUdp(netMethod int, h *Handler) {
	udpHandlers[netMethod] = h
}

func serveUDP(conn *net.UDPConn, remote *net.UDPAddr, data []byte, rlen int) {

	beginTime := time.Now()

	defer duxFramework.BufPond["UDP"].Put(data)
	buf := data[:rlen]

	if duxFramework.debug {
		glog.Infof("udp body : %v", base64.StdEncoding.EncodeToString(buf))
	}

	netMethod := GetNetMethod(buf)

	ctx := new(Context)
	ctx.Method = fmt.Sprintf("%v", netMethod)
	ctx.Proto = UDP
	ctx.Conn = conn
	ctx.RemoteAddr = remote.String()

	var err error
	var d []byte

	h := udpHandlers[netMethod]
	if h == nil {
		glog.Errorf("%v qhsec.NewServerUnpacker error : netMethod %v no handler.", remote, netMethod)
		return
	}
	d = buf

	//////////////////////////////////////////////////////////////
	/**
	var unpacker *qhsec.ServerUnpacker
	if h.cipherFlag&DecryptRequest > 0 {
		unpacker, err = qhsec.NewServerUnpacker(h.m.NppConfig())
		if err != nil {
			glog.Errorf("%v qhsec.NewServerUnpacker error : %v", remote, err.Error())
			return
		}
		defer unpacker.Close()

		dec, err := unpacker.Unpack(buf)
		if err != nil {
			glog.Errorf("%v qhsec.Unpack error : %v [%v]", remote, err.Error(), base64.StdEncoding.EncodeToString(buf))
			return
		}

		d = dec
		if duxFramework.debug {
			glog.Infof("http body decrypted : %v", base64.StdEncoding.EncodeToString(d))
		}
	}
	**/
	//////////////////////////////////////////////////////////////

	send, err := h.f(d, ctx)
	if err != nil {
		glog.Errorf("%v application callback handling error : %v", remote, err.Error())
		return
	}

	d = send

	/***
	if h.cipherFlag&EncryptResponse > 0 {
		d, err = unpacker.Pack(d)
		if err != nil {
			glog.Errorf("%v application qhsec.Pack error : %v", remote, err.Error())
			return
		}
	}
	***/

	conn.WriteToUDP(d, remote)

	if duxFramework.debug {
	    costMs := float64(time.Since(beginTime).Nanoseconds()) / 1000000.0
		glog.Infof("%v\tsend:%v\trecv:%v\tcost:%v", remote, len(buf), len(d), costMs)
	}
}

package unis

import (
	"encoding/base64"
	_"runtime"
	_"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/valyala/fasthttp"
//	"golib/cgo/qhsec"
)

func (h Handler) ServeHTTP(reqCtx *fasthttp.RequestCtx) {

    is_ok := DefaultFramework.check_client_ipf( reqCtx.RemoteAddr().String() )
    if !is_ok {
		reqCtx.SetStatusCode(579)
        reqCtx.Write([]byte("you have no access to ckproxy"))
        return
    }

	beginTime := time.Now()
	buf := reqCtx.PostBody()

	if duxFramework.debug {
		glog.Infof("http body : %v", base64.StdEncoding.EncodeToString(buf))
	}

	ctx := new(Context)
	ctx.Method = string(reqCtx.Path())
	ctx.Proto = HTTP
	ctx.Conn = nil
	ctx.HttpReq = reqCtx
	ctx.RemoteAddr = reqCtx.RemoteAddr().String()

	d := buf

	/////////////// decrypt request from client //////////////////////////
    /***
	var unpacker *qhsec.ServerUnpacker = nil
	if h.cipherFlag & DecryptRequest > 0 {
		unpacker, err := qhsec.NewServerUnpacker(h.m.NppConfig())
		if err != nil {
			glog.Errorf("%v %v?%v qhsec.NewServerUnpacker error : %v", reqCtx.RemoteAddr(), reqCtx.Path(),
																		reqCtx.URI().QueryString(), err.Error())
			reqCtx.SetStatusCode(500)
			return
		}
		defer unpacker.Close()

		dec, err := unpacker.Unpack(buf)
		if err != nil {
			glog.Errorf("%v %v?%v qhsec.Unpack error : %v [%v]", reqCtx.RemoteAddr(), reqCtx.Path(),
										reqCtx.URI().QueryString(), err.Error(), base64.StdEncoding.EncodeToString(buf) )
			reqCtx.SetStatusCode(400)
			return
		}
		d = dec

		if duxFramework.debug {
			glog.Infof("http body decrypted : %v", base64.StdEncoding.EncodeToString(d))
		}
	}
    ***/

	///////////////// proxy callback to 360 public cloud /////////////////
	send, err := h.f(d, ctx)
	if err != nil {
		glog.Errorf("%s %s?%s application callback handling error : %v",
					reqCtx.RemoteAddr().String(), reqCtx.Path(), reqCtx.URI().QueryString(), err.Error())

		reqCtx.SetStatusCode(500)
		return
	}
	d = send

	////////// encrypt response from 360 public cloud ////////////////////
    /***
	if h.cipherFlag & EncryptResponse > 0 {
		d, err = unpacker.Pack(d)
		if err != nil {
			glog.Errorf("%v %v?%v application qhsec.Pack error : %v",
							reqCtx.RemoteAddr().String(), reqCtx.Path(), reqCtx.URI().QueryString(), err.Error())
			reqCtx.SetStatusCode(500)
			return
		}
	}
    ***/
	//////////////////////////////////////////////////////////////////////
	reqCtx.Write(d)

	if duxFramework.debug {
	    costMs := float64(time.Since(beginTime).Nanoseconds()) / 1000000.0
		glog.Infof("%v\t%v\tsend:%v\trecv:%v\tcost:%v", reqCtx.RemoteAddr(), reqCtx.URI().String(), len(buf), len(d), costMs)
	}
}

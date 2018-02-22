package unis

import (
	"net"
	"strconv"
)

type HandlerFunc func(recv []byte, ctx *Context) (send []byte, err error)
type OnConnFunc func(lconn net.Conn)

type Handler struct {
	m          Module
	f          HandlerFunc
	cipherFlag int // Please see the defition of Cipher Flag, DecryptRequest/EncryptResponse/DefaultPlain/DefaultCipher

	onNewConn   OnConnFunc
	onCloseConn OnConnFunc
}

// Cipher Flag definition
const (
	DefaultPlain    int = 0                                // 明文协议默认标记(由应用层处理原始裸报文)
	DecryptRequest  int = 1                                // 客户端到服务器的请求数据，需要在框架解密后再调用应用层毁掉函数
	EncryptResponse int = 2                                // 服务器给客户端的响应数据，需要在框架加密后再发送出去
	DefaultCipher   int = DecryptRequest | EncryptResponse // 密文协议标记，由框架处理加密相关的细节问题，应用看到的都是明文。
)

/**
 *  HandleFunc 将 pattern 对应的处理函数直接注册到 fasthttp 框架中，从而省去了二次查找的过程
 * 	pattern -
 *     For HTTP, it is the URI of this request
 *     For UDP/TCP, it is an integer string indicating the net method id
 *	flag - 加解密相关的标记，参考 Cipher Flag definition
 */
func HandleFunc(pattern string, flag int, handler HandlerFunc, module Module) {

	h := &Handler{
		m:          module,
		f:          handler,
		cipherFlag: flag,
	}

    if pattern == "/api/system_repair.json" {
	    fasthttp_router.GET(pattern, h.ServeHTTP)
    }

    if pattern == "/ckproxy/status" {
	    fasthttp_router.GET(pattern, h.ServeHTTP)
    }

	fasthttp_router.POST(pattern, h.ServeHTTP)

	method, err := strconv.Atoi(pattern)
	if err == nil {
		handleTcp(method, h)
		handleUdp(method, h)
	}
}

func UDPHandleFunc(pattern string, flag int, handler HandlerFunc, module Module) {

	h := &Handler{
		m:          module,
		f:          handler,
		cipherFlag: flag,
	}

	method, err := strconv.Atoi(pattern)
	if err == nil {
		handleUdp(method, h)
	}
}

func HandleOnConnFunc(pattern string, onNewConnFunc OnConnFunc, onCloseConnFunc OnConnFunc) {
	method, err := strconv.Atoi(pattern)
	if err != nil {
		return
	}
	handleOnConn(method, onNewConnFunc, onCloseConnFunc)
}

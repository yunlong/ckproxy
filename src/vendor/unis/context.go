package unis

import (
	"fmt"
	"net"
	"unis/tcp"

	"github.com/valyala/fasthttp"
)

var _ = fmt.Print

// A NetProtocol represents the network protocol type of a client connection to a server.
type NetProtocol int

const (
	Unknown NetProtocol = iota + 1
	UDP
	TCP
	HTTP
)

type Context struct {
	// method is used to dispath this message to which function of the module to handle this message
	//     For HTTP, it is the URI of this request
	//     For UDP/TCP, it is an integer string indicating the net method id
	Method string

	Proto NetProtocol

	// The received data
	// for HTTP, it is the original body data of HTTP POST
	// for UDP, it is the original datagram package
	// for TCP, TODO
	recv []byte

	//TODO add cipher message
	HttpReq  *fasthttp.RequestCtx // or nil when not using HTTP

	Conn       net.Conn      // TCP or UDP connection with the client
	RemoteAddr string

	TCPAccessor tcp.TCPAccessor
}

//func (c* Context) GetExtra(id int) ([]byte, error) {
//	if b, ok := c.extras[id]; ok {
//		return b, nil
//	}
//	return nil, fmt.Errorf("Not found extra id %v", id)
//}
//
//func (c* Context) GetExtras() tcp.Extras {
//	return c.extras
//}
//

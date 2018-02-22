package unis

import (
	"github.com/golang/glog"
	"github.com/valyala/fasthttp"
)

type HttpContext struct {
	ctx *fasthttp.RequestCtx
	h Handler
}

type Dispatcher struct {
	c             chan *HttpContext
	maxProcessNum int
	maxQueueNum   int
}

func NewDispatcher(maxProcessNum, maxQueueNum int) *Dispatcher {

	d := &Dispatcher{
		c:             make(chan *HttpContext, maxQueueNum),
		maxProcessNum: maxProcessNum,
		maxQueueNum:   maxQueueNum,
	}

	d.Init()

	return d

}

func (d *Dispatcher) Init() {

	for i := 0; i < d.maxProcessNum; i++ {
		go ServeHTTPLimited(d.c)
	}
}

func (d *Dispatcher) Close() {
	close(d.c)
}

func (d *Dispatcher) Dispatch(c *HttpContext) {
	d.c <- c
}

func ServeHTTPLimited(c chan *HttpContext) {

	for {

		httpContext, ok := <-c
		if !ok || httpContext == nil {
			glog.Errorf(" http channel is broken or quited")
			continue
		}

		httpContext.h.ServeHTTP(httpContext.ctx)
	}
}

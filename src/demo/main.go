package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

/*
	Test commands after the server startup :

	$ curl http://localhost:9360/demoecho -d xxdddfaxyss
	$ curl http://localhost:9360/demoproxy?u=http://360.cn
	$ curl http://localhost:9360/status.html
*/

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

var host = flag.String("addr", "0.0.0.0:9080", "host:port")

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	flag.Parse()
	srv := http.Server{Handler: &Handler{}}
	listener, err := net.Listen("tcp", *host)
	if err != nil {
		panic("http listen error" + err.Error())
	}

	err = srv.Serve(listener)

	if err != nil {
		panic("http listen error" + err.Error())
	}

}

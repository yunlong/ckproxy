package main

import (
	"unis"
	"unis/sampleapp/demo"
)

/*
	Test commands after the server startup :

	$ curl http://localhost:9360/demoecho -d xxdddfaxyss
	$ curl http://localhost:9360/demoproxy?u=http://360.cn
	$ curl http://localhost:9360/status.html
*/

func main() {
	fw := unis.DefaultFramework
	fw.RegisterModule("demoproxy", new(demo.DemoModule))
	fw.RegisterModule("lddtcp", new(demo.LddTcpModule))
	err := fw.Initialize()
	if err != nil {
		panic(err.Error())
	}

	fw.Run()
}

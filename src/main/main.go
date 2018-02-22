package main

import (
	"fmt"
	"time"
	"os"
   _ "net/http/pprof"
    "runtime"
    "unis"
    "ckproxy"
)

var ckproxy_banner string = `
      _
  ___| | ___ __  _ __ _____  ___   _
 / __| |/ / '_ \| '__/ _ \ \/ / | | |
| (__|   <| |_) | | | (_) >  <| |_| |
 \___|_|\_\ .__/|_|  \___/_/\_\\__, |
          |_|                  |___/
`

func main() {

	fmt.Print(ckproxy_banner)
   	fmt.Printf("Welcome to the ckproxy\n")
   	fmt.Printf("ckproxy started at: %s\n", time.Now().Format("2006-01-02 15:04:05"))

    defer func() {
        if err := recover(); err != nil {
            var st = func(all bool) string {
                // Reserve 1K buffer at first
                buf := make([]byte, 512)

                for {
                    size := runtime.Stack(buf, all)
                    // The size of the buffer may be not enough to hold the stacktrace,
                    // so double the buffer size
                    if size == len(buf) {
                        buf = make([]byte, len(buf)<<1)
                        continue
                    }
                    break
                }

                return string(buf)
            }
            fmt.Printf("panic: %s\n", err)
            fmt.Printf("stack: %s\n", st(false))
        }
    }()

	fw := unis.DefaultFramework
	fw.RegisterModule("ckproxy", new(ckproxy.CKProxy))
	err := fw.Initialize()
	if err != nil {
   		fmt.Printf("ckproxy initialize failed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		os.Exit(0)
	}

	fw.Run()

    os.Exit(0)
}

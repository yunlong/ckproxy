// +build ignore

package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("addr", "127.0.0.1:8000", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	c.SetPingHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		log.Printf("receiving PingMessage from server %s\n", c.RemoteAddr())

		c.SetWriteDeadline(time.Now().Add(60 * time.Second))
		if err := c.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
			return nil
		}

		log.Printf("sending PongMessage to peer %s\n", c.RemoteAddr())
		return nil
	})
	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("receiving TextMessage %s from server %s\n", message, c.RemoteAddr())
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.Format("2006-01-02 15:04:05")+" \"Hello Server\""))
			if err != nil {
				log.Println("write:", err)
				return
			}
			log.Printf("sending TextMessage %s to server %s\n", t.Format("2006-01-02 15:04:05")+" \"Hello Server\"", c.RemoteAddr())

		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

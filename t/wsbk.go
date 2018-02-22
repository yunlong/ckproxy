/**
 * @author         YunLong.Lee    <yunlong.lee@163.com>
 * @version        0.5
 */
package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
	"encoding/binary"
	"crypto/rand"
	"fmt"

	"github.com/golang/glog"
	"github.com/cheggaaa/pb"
)

var (
	addr 	= flag.String("addr", "127.0.0.1:8000", "service address")
	conn 	= flag.Int("conn", 10, "num of connection")
	ack_rto = flag.Bool("rto", false, "timeout retransmission")
)

func countMsg() {
	for {
		select {
		case <-is_have_msg:
			recv_msg_count++
		}
	}
}

func send_ack_msg_to_server( c *websocket.Conn, ack_seq uint32 ) error {
	write_buf := make([]byte, 5)
	binary.BigEndian.PutUint32(write_buf, uint32(ack_seq) )
	b4 :=  byte(0x1 << 4 | 0x3 << 1 | 0x0)
	write_buf[4] = b4

	if err := c.WriteMessage(websocket.TextMessage, write_buf); err != nil {
		glog.Errorf("write:%s", err)
		return err
	}

	return nil
}

func send_notify_msg_to_server( c *websocket.Conn, ack_seq uint32 ) error {

	write_buf := make([]byte, 9 + len("Hello SkyTime"))
	binary.BigEndian.PutUint32(write_buf, uint32(ack_seq) )
	b4 :=  byte(0x1 << 4 | 0x2 << 1 | 0x0)
	write_buf[4] = b4

	body_len := len("Hello SkyTime")
	binary.BigEndian.PutUint32(write_buf[5:], uint32(body_len))
	copy(write_buf[9:], "Hello SkyTime")

	if err := c.WriteMessage(websocket.TextMessage, write_buf); err != nil {
		glog.Errorf("write:%s", err)
		return err
	}

	return nil
}

func benchmark(strUrl string) {

	defer wg.Done()

    skb := make([]byte, 16)
    rand.Read(skb)
	peer_id := []string{ fmt.Sprintf("%X-%X-%X-%X-%X", skb[0:4], skb[4:6], skb[6:8], skb[8:10], skb[10:]) }
	reqHdr := make(http.Header)
	reqHdr["peer_id"] = peer_id

	retry_cnt := 0

retry_conn:
	c, _, err := websocket.DefaultDialer.Dial(strUrl, reqHdr)

	if err != nil {
		glog.Error("dial:%s", err)

		time.Sleep(3 * time.Second);

		if retry_cnt < 5 {
			retry_cnt++
			goto retry_conn
		} else {
			return
		}
	}

	defer c.Close()

	done := make(chan struct{})

	c.SetPingHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(60 * time.Second))
		//	glog.Infof("receiving PingMessage from server %s\n", c.RemoteAddr())
		if err := c.WriteMessage(websocket.PongMessage, []byte{}); err != nil {
			return nil
		}
		//  glog.Infof("peer %s sending PongMessage to server %s\n", c.LocalAddr(), c.RemoteAddr())

		return nil
	})

	defer close(done)

	go func() {
		defer c.Close()
		defer close(done)
		for {
		    c.SetReadLimit(16384)
			c.SetReadDeadline(time.Now().Add(60 * time.Second))
			mt, message, err := c.ReadMessage()
			if err != nil {
				glog.Error("read:%s", err)
				return
			}

			if mt == websocket.TextMessage {

				if *ack_rto {

					ack_seq := binary.BigEndian.Uint32(message[0:4])
					msg_ver  := byte( message[4] >> 4 )
					msg_type := byte( message[4]  & 0xf  >> 1 )
					svr_flag := byte( message[4]  & 0x1 )

					if msg_type == 3 {

		                // recving ack message from server
		                ack_seq := binary.BigEndian.Uint32(message[0:4])
		                fmt.Printf("receiving ACK_SEQ %d msg_ver %d msg_type %d svr_flag %d from server %s \n",
				                        ack_seq, msg_ver, msg_type, svr_flag, c.RemoteAddr())
		           		// c.recv_ack_seq <- int( ack_seq )

					} else {

						body_len := binary.BigEndian.Uint32(message[5:9])

                        if msg_type == 1 {

						    glog.Infof("peer %s receiving Message \"%s\" " +
							                "with ack_seq %d msg_ver %d msg_type %d svr_flag %d body_len %d from server %s\n",
							            c.LocalAddr(), message[9:],
							            ack_seq, msg_ver, msg_type, svr_flag, body_len, c.RemoteAddr() )

						    glog.Infof("sending Message with ack_seq %d to %s \n", ack_seq, c.RemoteAddr())
						    send_ack_msg_to_server( c, ack_seq )
                        }

						glog.Infof("sending Notify Message with ack_seq %d to %s \n", ack_seq, c.RemoteAddr())
						send_notify_msg_to_server( c, ack_seq )

						is_have_msg <- true
					}

				} else {

		//	fmt.Printf("peer %s receiving Message \"%s\" from server %s\n", c.LocalAddr(), message, c.RemoteAddr())
					is_have_msg <- true

				}
			}
		}
	}()

}

func broadcastMessage(message string) {

	msg := message

	body := bytes.NewBuffer([]byte(msg))
	u := url.URL{Scheme: "http", Host: *addr, Path: "/push"}
	resp, err := http.Post(u.String(), "application/x-www-form-urlencoded", body)
	if err != nil {
		glog.Error(err)
	}

	defer resp.Body.Close()
	nbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
	}
	glog.Infof("%s", nbody)
}

func heartbeat(ch <-chan time.Time) {
	for t := range ch {
		//fmt.Printf("received %d broad message in %s\n", recv_msg_count, t.Format("2006-01-02 15:04:05"))
		glog.Infof("received %d broad message in %s\n", recv_msg_count, t.Format("2006-01-02 15:04:05"))
	}
}

var (
	interrupt      chan os.Signal
	wg             sync.WaitGroup
	is_have_msg    = make(chan bool)
	recv_msg_count int64
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Set("logtostderr", "true")
//	flag.Set("log_dir", "./log/")
	flag.Set("v", "3")

	flag.Parse()

	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	glog.Infof("connecting to %s", u.String())

	go countMsg()

	bar := pb.StartNew(*conn)
	for i := 0; i < *conn; i++ {
		bar.Increment()
		wg.Add(1)
		go benchmark(u.String())
		time.Sleep(5 * time.Millisecond)
	}
	bar.FinishPrint("register connection end .....")

	//////////////////// handler for heartbeat /////////////
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	go heartbeat(ticker.C)
	////////////////////////////////////////////////////////

	wg.Wait()
	
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {

		 select {

	//	case t := <-ticker.C:
		case <-ticker.C:

		//	glog.Infof("sending TextMessage %s\n", t.Format("2006-01-02 15:04:05") + " Hello Server ")

		case <-interrupt:

			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				glog.Errorf("write close: %v", err)
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

	glog.Flush()
	os.Exit(0)
}

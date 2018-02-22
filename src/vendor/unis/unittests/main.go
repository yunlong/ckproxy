package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"golib/cgo/qhsec"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
	"unis"
	"unis/sampleapp/demo"
)

/*
	Test commands after the server startup :

	$ curl http://localhost:9360/demoecho -d xxdddfaxyss
	$ curl http://localhost:9360/demoproxy?u=http://360.cn
	$ curl http://localhost:9360/status.html
*/

var exitCode = 0

type ExitModule struct{}

func (m *ExitModule) Initialize() error {
	println("ExitModule initializing ...")
	unis.HandleFunc("/exit", unis.DefaultPlain, m.Exit, m)
	return nil
}

func (m *ExitModule) NppConfig() *qhsec.NppConfig {
	return nil
}

func (m *ExitModule) Exit(recv []byte, ctx *unis.Context) (send []byte, err error) {
	println("server exiting with code : ", exitCode)
	os.Exit(exitCode)
	return recv, nil
}

var demoModule = new(demo.DemoModule)
var udpModule = new(demo.UdpModule)

func main() {
	// reset the default value of ConfPath
	flag.StringVar(unis.ConfPath, "ConfPath", "../conf/ut.ini", "The config file of unit test")

	fw := unis.DefaultFramework
	fw.RegisterModule("ExitModule", new(ExitModule))
	fw.RegisterModule("demoproxy", demoModule)
	fw.RegisterModule("udpModule", udpModule)
	err := fw.Initialize()
	if err != nil {
		panic(err.Error())
	}

	go fw.Run()

	time.Sleep(1 * time.Second)

	RunAllTests()
	Exit()
	os.Exit(exitCode)
}

func RunAllTests() {
	TestStatus()
	TestUdpCipher()
	TestCipherEcho()
}

func Exit() {
	url := "http://127.0.0.1:19361/exit"
	_, err := http.Get(url)
	if err != nil {
		exitCode = 1
		println("http.Get url ", url, " failed : ", err.Error())
		return
	}
}

func TestStatus() {
	url := "http://127.0.0.1:19361/status.html"
	resp, err := http.Get(url)
	if err != nil {
		exitCode = 1
		println("http.Get url ", url, " failed : ", err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil || string(body) != "OK" {
		exitCode = 1
	}
}

func TestCipherEcho() {
	nc := demoModule.NppConfig()
	if nc == nil {
		println("NppConfig ERROR")
		exitCode = 2
		return
	}
	packer, err := qhsec.NewClientPacker(nc, 11)
	if err != nil {
		println("create client packer failed", err.Error())
		exitCode = 3
		return
	}
	packer.SetOption(qhsec.OptSymmetricKeyNo, 8628)
	packer.SetOption(qhsec.OptSymmetricMethod, 2)
	hello := "hello"
	cipher, err := packer.Pack([]byte(hello))
	if err != nil {
		println("pack data failed", err.Error())
		exitCode = 4
		return
	}
	br := bytes.NewReader(cipher)
	url := "http://127.0.0.1:19361/democipherecho"
	resp, err := http.Post(url, "image/jpeg", br)
	if err != nil {
		exitCode = 1
		println("http.Get url ", url, " failed : ", err.Error())
		return
	}

	defer resp.Body.Close()
	cipher, err = ioutil.ReadAll(resp.Body)
	println("recv data : ", base64.StdEncoding.EncodeToString(cipher), "")
	body, err := packer.Unpack(cipher)
	if err != nil || string(body) != hello {
		exitCode = 1
	}
}

func TestUdpCipher() {

	nc := udpModule.NppConfig()
	if nc == nil {
		println("NppConfig ERROR")
		exitCode = 2
		return
	}
	packer, err := qhsec.NewClientPacker(nc, 11)
	if err != nil {
		println("create client packer failed", err.Error())
		exitCode = 3
		return
	}
	packer.SetOption(qhsec.OptSymmetricKeyNo, 8628)
	packer.SetOption(qhsec.OptSymmetricMethod, 2)
	packer.SetOption(qhsec.OptNetMethod, 3)
	hello := "hello"
	cipher, err := packer.Pack([]byte(hello))
	if err != nil {
		println("pack data failed", err.Error())
		exitCode = 4
		return
	}

	addr, err := net.ResolveUDPAddr("udp", ":19362")
	socket, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		println("连接失败!", err)
		return
	}
	defer socket.Close()

	_, err = socket.Write(cipher)
	if err != nil {
		println("发送数据失败!", err)
		return
	}

	data := make([]byte, 65536)
	rlen, _, err := socket.ReadFromUDP(data)
	if err != nil {
		println("读取数据失败!", err)
		return
	}

	body, err := packer.Unpack(data[:rlen])
	if err != nil || string(body) != hello {
		println("udp unpack error")
		exitCode = 1
	}

}

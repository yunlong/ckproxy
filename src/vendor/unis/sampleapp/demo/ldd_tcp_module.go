package demo

import (
	"encoding/base64"
	"fmt"
	"golib/cgo/qhsec"
	"unis"
)

type LddTcpModule struct {
	nppConfig *qhsec.NppConfig
}

func (m *LddTcpModule) Initialize() error {

	println("LddTcpModule initializing ...")
	unis.HandleFunc("1100", unis.DefaultPlain, m.TcpEcho, m)
	//unis.HandleFunc("1100", unis.DefaultCipher, m.CipherTcpEcho, m)
	
	err := m.InitNppConfig()
	if err != nil {
		return err
	}
	return nil

}

func (m *LddTcpModule) NppConfig() *qhsec.NppConfig {
	return m.nppConfig
}

func (m *LddTcpModule) InitNppConfig() error {

	fw := unis.DefaultFramework
	asymmetric_key_path := fw.GetPathConfig("demo", "asymmetric_keys_path")
	symmetric_keys := fw.GetPathConfig("demo", "symmetric_keys_path")
	business_name, _ := fw.Conf.SectionGet("demo", "business_name")
	p, err := qhsec.NewNppConfig(symmetric_keys, business_name, asymmetric_key_path)
	if err != nil {
		return err
	}
	m.nppConfig = p
	return nil
}

func (m *LddTcpModule) TcpEcho(recv []byte, ctx *unis.Context) (send []byte, err error) {

	fmt.Printf("url=%v data=[%v]\n", ctx.Method, base64.StdEncoding.EncodeToString(recv))
	send = []byte("err=0\r\n0\t\r\n")
	return send, nil

}

func (m *LddTcpModule) CipherTcpEcho(recv []byte, ctx *unis.Context) (send []byte, err error) {

	fmt.Printf("url=%v post data=[%v]\n", ctx.Method, base64.StdEncoding.EncodeToString(recv))
	send = []byte("err=0\r\n0\t\r\n")
	return send, err

}

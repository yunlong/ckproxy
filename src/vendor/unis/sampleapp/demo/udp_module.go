package demo

import (
	"fmt"
	"golib/cgo/qhsec"
	"unis"
)

type UdpModule struct {
	nppConfig *qhsec.NppConfig
}

func (m *UdpModule) Initialize() error {
	println("UdpModule initializing ...")
	//	unis.HandleFunc("2", unis.DefaultPlain, m.UdpEcho, m)
	unis.HandleFunc("3", unis.DefaultCipher, m.CipherUdpEcho, m)
	err := m.InitNppConfig()
	if err != nil {
		return err
	}
	return nil
}

func (m *UdpModule) NppConfig() *qhsec.NppConfig {
	return m.nppConfig
}

func (m *UdpModule) InitNppConfig() error {
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

func (m *UdpModule) CipherUdpEcho(recv []byte, ctx *unis.Context) (send []byte, err error) {
	fmt.Printf("post data=[%v]\n", string(recv))
	send = recv
	return send, err
}

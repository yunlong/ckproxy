package ldd

import (
	"errors"
	"fmt"
	"net"
	_ "time"
)

type LCSClient struct {
	lt   *LddTCP
	conn *net.TCPConn
	Host string
}

//const (
//	// 请求方向
//	kRequest = 0
//	kPing    = 1
//	kCancel  = 2
//
//	// 响应方向
//	kResponse = 128
//	kPong     = 129
//	kLast     = 130
//	kEnd      = 131
//)

const (
	kReqeustVersion = 0xA1
)

func (client *LCSClient) GetPayloadType() (payloadType uint16) {
	return client.lt.PayloadType()
}

func (client *LCSClient) Write(ldd_type int, payload []byte) (err error) {
	client.lt.Type = kRequest
	client.lt.Version = kReqeustVersion
	client.lt.IncraseSeq()
	if client.conn == nil {
		return errors.New("invalid connection")
	}
	return client.lt.WriteEx(payload, client.conn, ldd_type, true)
}

func (client *LCSClient) Read() (payload []byte, err error) {
	if client.conn == nil {
		return nil, errors.New("invalid connection")
	}
	err = client.lt.ReadHeader(client.conn)
	if err != nil {
		return payload, err
	}

	err = client.lt.ReadBody(client.conn)

	if err != nil {
		return payload, err
	}

	payload = client.lt.Payload()
	return payload, nil
}

func (client *LCSClient) PingPong() error {
	if client.conn == nil {
		return errors.New("invalid connection")
	}
	client.lt.Type = kPing
	client.lt.Version = kReqeustVersion
	client.lt.IncraseSeq()
	err := client.lt.WriteEx(nil, client.conn, 0, true)
	if err != nil {
		return err
	}

	err = client.lt.ReadHeader(client.conn)
	return err
}

func (client *LCSClient) ParseFromBuf(buf []byte) (payload []byte, err error) {

	err = client.lt.ReadHeaderFromBuf(buf)
	if err != nil {
		return payload, err
	}

	err = client.lt.ReadBodyFromBuf(buf)

	if err != nil {
		return payload, err
	}

	payload = client.lt.Payload()

	return payload, nil
}

func (client *LCSClient) Fetch(ldd_type int, payload []byte) (ret_payload []byte, err error) {
	err = client.Write(ldd_type, payload)
	if err != nil {
		fmt.Printf("Fetch write failed.payload:%v, err:%v", payload, err)
		return ret_payload, err
	}
	ret_payload, err = client.Read()
	if err != nil {
		fmt.Printf("Fetch read failed. conn:%v", client.conn)
	}
	return ret_payload, err
}

func (client *LCSClient) Reconnect() (err error) {
	if client.conn != nil {
		client.conn.Close()
	}
	client.conn, err = NewConn(client.Host)
	return err
}

func (client *LCSClient) Close() {
	if client.conn != nil {
		client.conn.Close()
	}
}

func NewClient(addr string) (client *LCSClient, err error) {
	client = &LCSClient{
		Host: addr,
	}
	client.lt = MakeLddTCP()
	client.conn, err = NewConn(addr)
	return client, err
}

func NewConn(host string) (conn *net.TCPConn, err error) {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return nil, err
	}

	conn, err = net.DialTCP("tcp", nil, addr)
	return conn, err
}

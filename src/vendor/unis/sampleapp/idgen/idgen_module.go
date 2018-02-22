package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"unis"

	"golib/cgo/qhsec"
)

type IdGenModule struct {
	id    uint64
	mutex sync.Mutex
}

type Response struct {
	StartIndex uint64 `json:"start_index"`
	Count      int    `json:"count"`
}

func (m *IdGenModule) Initialize() error {
	fw := unis.DefaultFramework
	fw.Logger.Info("IdGenModule", "IdGenModule initializing ...")
	unis.HandleFunc("/idgen", unis.DefaultPlain, m.GenerateId, m)

	if c, ok := fw.Conf.SectionGetInt("idgen", "start_index"); ok {
		m.id = uint64(c)
	}
	fw.Logger.Info("IdGenModule", "IdGenModule start_index=%v", m.id)
	return nil
}

func (m *IdGenModule) NppConfig() *qhsec.NppConfig {
	return nil
}

func (m *IdGenModule) GenerateId(recv []byte, ctx *unis.Context) (send []byte, err error) {
	r := ctx.HttpReq
	r.ParseForm()
	count := r.FormValue("count")
	if len(count) == 0 {
		return []byte(""), fmt.Errorf("cannot find 'count' parameter")
	}
	c, err := strconv.Atoi(count)
	if err != nil {
		return []byte(""), fmt.Errorf("convert 'count' to integer failed.")
	}

	var response Response
	response.Count = c
	m.mutex.Lock()
	response.StartIndex = m.id
	m.id += uint64(c)
	m.mutex.Unlock()
	send, err = json.Marshal(&response)
	unis.DefaultFramework.Logger.Info("IdGenModule", "response:%v   [%v]", response, string(send))
	return send, err
}

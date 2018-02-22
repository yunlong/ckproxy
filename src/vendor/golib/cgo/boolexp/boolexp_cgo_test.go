package boolexp

import (
	"testing"
)

func TestBoolExp(t *testing.T) {
	b, err := NewBoolExp(string("%fver_360safe.exe% hasAny \"1.0.0.0\""))

	if err != nil {
		t.Errorf("test erorr, init error %v %v", err, b)
	}

	req := &Request{
		Asks: []*Ask{},
	}

	ask := &Ask{
		Conditions: make(map[string]string),
	}

	ask.Conditions["fver_360safe.exe"] = "1.0.0.0,2.1.0.0"
	ask.Conditions["fver_360"] = "1.0.0.0,2.1.0.0"
	req.Asks = append(req.Asks, ask)

	resp, err := b.Process(req)
	if err != nil {
		t.Errorf("process error:%v", err, resp)
	}

	//t.Errorf("process error:%v", err, resp)
	for k, v := range resp.Anss {
		t.Log("%v %v", k, v)
	}

	if resp.Anss[0].Result != true {
		t.Errorf("get result not true")
	}

	b.Close()
}

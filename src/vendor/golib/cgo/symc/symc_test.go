package symc

import (
	"fmt"
	"testing"
	"time"
)

func TestSymcCluster(t *testing.T) {
	s, err := NewSymc(string("etc/vbucket_conf.ini"), "symc_ycs_docid_cgo")
	if err != nil {
		t.Errorf("test erorr, init error %v %v", err, s)
		return
	}
	defer s.Close()
	kv := map[string]string{
		"160707650092850":  "",
		"160707650092877":  "",
		"160707650092873":  "",
		"160707650092885":  "",
		"160707650092949":  "",
		"1607076500929494": "",
		"1607076500929493": "",
		"1607076500929492": "",
		"1607076500929491": "",
		"1607076500929495": "",
		"1607076500929496": "",
		"1607076500929497": "",
		"1607076500929498": "",
		"1160707650092949": "",
	}

	kvResult, ret := s.Get(kv)
	if ret != nil {
		t.Errorf("%v %v", kvResult, ret)
	}
}

func TestSymcCluster2(t *testing.T) {
	s, err := NewSymc(string("etc/vbucket_conf.ini"), "symc_ycs_docid_cgo")
	if err != nil {
		t.Errorf("test erorr, init error %v %v", err, s)
		return
	}
	defer s.Close()
	kv := map[string]string{
		"NULL_160707650092850":  "",
		"NULL_160707650092877":  "",
		"NULL_160707650092873":  "",
		"NULL_160707650092885":  "",
		"NULL_160707650092949":  "",
		"NULL_1607076500929494": "",
		"NULL_1607076500929493": "",
		"NULL_1607076500929492": "",
		"NULL_1607076500929491": "",
		"NULL_1607076500929495": "",
		"NULL_1607076500929496": "",
		"NULL_1607076500929497": "",
		"NULL_1607076500929498": "",
		"NULL_1160707650092949": "",
	}

	kvResult, ret := s.Get(kv)
	if ret != nil {
		t.Errorf("%v %v", kvResult, ret)
	}
}

func testSymcClose(t *testing.T) {

	s, err := NewSymc(string("etc/vbucket_conf.ini"), "symc_cgo_test")
	if err != nil {
		t.Errorf("test erorr, init error %v %v", err, s)
		return
	}
	_ = s
	defer s.Close()
	kvQuery := map[string]string{
		"symc_1": "",
		"symc_2": "",
		"symc_3": "",
		"symc_4": "",
		"symc_5": "",
	}
	kvResult, ret := s.Get(kvQuery)
	if ret != nil {
		t.Errorf("%v %v", kvResult, ret)
	}

}
func TestSymcClose(t *testing.T) {
	for i := 0; i < 2000; i++ {
		testSymcClose(t)
	}
}
func TestSymc(t *testing.T) {
	s, err := NewSymc(string("etc/vbucket_conf.ini"), "symc_cgo_test")

	if err != nil {
		t.Errorf("test erorr, init error %v %v", err, s)
		return
	}
	_ = s

	kv := map[string]string{
		"symc_1": "11111",
		"symc_2": "2",
		"symc_3": "3",
		"symc_4": "41199999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999911",
		"symc_5": "5",
	}

	kvQuery := map[string]string{
		"symc_1": "",
		"symc_2": "",
		"symc_3": "",
		"symc_4": "",
		"symc_5": "",
	}

	for i := 0; i < 200; i++ {
		id := fmt.Sprintf("go_test%d", i)
		value := fmt.Sprintf("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx%d", i)
		kv[id] = value
		kvQuery[id] = ""
	}

	ret := s.Set(kv)
	t.Logf("%v", ret)
	if ret != nil {
		t.Errorf("%v", ret)
	}
	kvResult, ret := s.Get(kvQuery)
	if ret != nil {
		t.Errorf("%v %v", kvResult, ret)
	}

	if kvResult["symc_1"] != kv["symc_1"] {
		t.Errorf("%v %v error", kvResult, kv)
	}

	kv2 := map[string]string{
		"xxxxxx": "",
		"xxxxx1": "",
		"xxxxx2": "",
		"xxxxx3": "",
	}
	kvResult, ret = s.Get(kv2)
	kvResult, ret = s.Get(kv2)
	kvResult, ret = s.Get(kv2)
	kvResult, ret = s.Get(kv2)
	kvResult, ret = s.Get(kv2)
	kvResult, ret = s.Get(kv2)
	for i := 0; i < 20000; i++ {
		kvResult, ret = s.Get(kv2)
	}

	s.Close()

	//t.Errorf("end")
}

func testSymcPool(p *Pool, t *testing.T, kv, result map[string]string) {
	s, err := p.GetSymc()
	if err != nil {
		t.Errorf("test erorr, get error %v %v", err, s)
		return
	}
	defer s.Close()
	kvResult, ret := s.Get(kv)
	if ret != nil {
		t.Errorf("%v %v", kvResult, ret)
	}
	for k, v := range result {
		if v != kvResult[k] {
			t.Errorf("%v %v", kvResult, result)
		}
	}

	kv2 := map[string]string{
		"xxxxxx": "",
		"xxxxx1": "",
		"xxxxx2": "",
		"xxxxx3": "",
	}
	kvResult, ret = s.Get(kv2)
	for i := 0; i < 10000; i++ {
		kvResult, ret = s.Get(kv2)
	}
	t.Logf("ok")
}
func TestSymcPool(t *testing.T) {
	p, err := NewSymcPool(string("etc/vbucket_conf.ini"), "symc_cgo_test")
	defer p.Close()

	if err != nil {
		t.Errorf("test erorr, init error %v %v", err, p)
	}

	kv := map[string]string{
		"symc_1": "11111",
		"symc_2": "2",
		"symc_3": "3",
		"symc_4": "41199999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999911",
		"symc_5": "5",
	}

	kvQuery := map[string]string{
		"symc_1": "",
		"symc_2": "",
		"symc_3": "",
		"symc_4": "",
		"symc_5": "",
	}

	for i := 0; i < 100; i++ {
		go testSymcPool(p, t, kvQuery, kv)
	}
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	time.Sleep(10 * time.Microsecond)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	time.Sleep(10 * time.Microsecond)
	go testSymcPool(p, t, kvQuery, kv)
	go testSymcPool(p, t, kvQuery, kv)
	testSymcPool(p, t, kvQuery, kv)

	time.Sleep(30 * time.Second)

	//t.Errorf("end")
}

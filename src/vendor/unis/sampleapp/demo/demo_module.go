package demo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golib/cgo/qhsec"
	"golib/dbuf"
	"unis"
)

type DemoModule struct {
	nppConfig *qhsec.NppConfig
	dict      *dbuf.DoubleBuffering
}

func (m *DemoModule) Initialize() error {
	println("DemoModule initializing ...")
	unis.HandleFunc("/demoproxy", unis.DefaultPlain, m.Proxy, m)
	unis.HandleFunc("/demoecho", unis.DefaultPlain, m.Echo, m)
	unis.HandleFunc("/democipherecho", unis.DefaultCipher, m.Echo, m)
	unis.HandleFunc("/dict", unis.DefaultPlain, m.SearchDict, m)
	err := m.InitNppConfig()
	if err != nil {
		return err
	}

	name := "mydict"
	fw := unis.DefaultFramework
	rc := fw.DoubleBufferingManager.Add(name, "the config data of Dict or the config file path of Dict", newDict)
	if rc == false {
		return errors.New("Dict initialize failed")
	}
	m.dict = fw.DoubleBufferingManager.Get(name)
	return nil
}

func (m *DemoModule) NppConfig() *qhsec.NppConfig {
	return m.nppConfig
}

func (m *DemoModule) InitNppConfig() error {
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

func (m *DemoModule) Proxy(recv []byte, ctx *unis.Context) (send []byte, err error) {
	fmt.Printf("url=%v post data=[%v] querys=%v\n", ctx.Method, string(recv), ctx.HttpReq.URL.Query())
	proxyurls, ok := ctx.HttpReq.URL.Query()["u"]
	if !ok || len(proxyurls) == 0 {
		return []byte("not found proxy url by \"u\""), nil
	}

	proxyurl := proxyurls[0]
	resp, err := http.Get(proxyurl)
	if err != nil {
		return []byte(fmt.Sprintf("http.Get(%v) failed : %v", proxyurl, err.Error())), nil
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (m *DemoModule) Echo(recv []byte, ctx *unis.Context) (send []byte, err error) {
	//fmt.Printf("url=%v post data=[%v] querys=%v\n", ctx.Method, string(recv), ctx.HttpReq.URL.Query())
	return recv, nil
}

func (m *DemoModule) SearchDict(recv []byte, ctx *unis.Context) (send []byte, err error) {
	t := m.dict.Get()
	if t.Target == nil {
		return []byte("ERROR, DoubleBuffering.Get return nil"), nil
	}
	defer t.Release()        // 注意这个语句，必须调用。类似于 github.com/garyburd/redigo/redis 里面的 redis.Pool 使用方法。
	dict := t.Target.(*Dict) // 转换为具体的Dict对象
	if dict == nil {
		return []byte("ERROR, Convert DoubleBufferingTarget to Dict failed"), nil
	}

	return []byte(dict.d), nil
}

////////////////////////
// Dict 实现了 dbuf.DoubleBufferingTarget 接口
type Dict struct {
	d string
	//业务自己的其他更复杂的数据结构
}

func newDict() dbuf.DoubleBufferingTarget {
	d := new(Dict)
	return d
}

/*
请求： curl http://localhost:9360/dict
Reload指令：curl "http://localhost:9360/admin/reload?name=mydict&path=xxx2342c"
*/
func (d *Dict) Initialize(conf string) bool {
	// 这个conf一般情况下是一个配置文件的路径
	// 这里我们简单的认为它只是一段数据
	d.d = conf
	return true
}

func (d *Dict) Close() {
	// 在这里做一些资源释放工作
	// 当前的这个示例代码没有资源需要释放，就留空
	fmt.Printf("calling Dict.Close() ...\n")
}

package hadoop_sconf

/*
#cgo CXXFLAGS:		 -I. -I../include -I../../include -I/home/s/safe/hadoop_sconf
#cgo CXXFLAGS:		 -I/home/s/safe/include
#cgo CXXFLAGS:		 -I/home/s/include
#cgo LDFLAGS:		 -L/home/s/safe/lib
#cgo LDFLAGS:		 -L/home/s/lib/
#cgo LDFLAGS:		 -lstdc++ -lcurl
#cgo LDFLAGS:		 -lpthread -ldl -lnetproto
#cgo LDFLAGS:		 /home/s/safe/hadoop_sconf/libhadoop_sconf.a
#cgo LDFLAGS:       -Wl,-rpath=/home/s/safe/protobuf3/lib/
#cgo LDFLAGS:       -Wl,-rpath=/home/s/safe/lib
#cgo LDFLAGS:       -Wl,-rpath=/home/s/lib

#include <stdlib.h>
#include "hadoop_sconf.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

type KeyPair struct {
	Id    string
	Key   string
	idInt int
}

type HadoopSconf struct {
	h       C.hadoop_sconf_result_t
	keys    map[string]*KeyPair
	keylist []*KeyPair
	pos     int
	size    int
}

func NewHadoopSconf(url string) (*HadoopSconf, error) {
	h := &HadoopSconf{
		keys:    map[string]*KeyPair{},
		keylist: []*KeyPair{},
		pos:     0,
	}

	c_url := C.CString(url)
	defer C.free(unsafe.Pointer(c_url))
	h.h = C.raw_do_hadoop_sconf_post(c_url)
	if h.h == nil {
		return nil, errors.New("create sconf fail")
	}

	keystr := h.getData()
	err := h.getKeys(&keystr)
	return h, err
}

func (h *HadoopSconf) getData() string {
	//var l C.int
	//l = C.get_data_len(h.h)
	//buffer := make([]byte, int(l))
	cbuff := C.get_data(h.h)
	buff := C.GoString(cbuff)
	return buff
}

func (h *HadoopSconf) getKeys(str *string) error {
	li := strings.Split(*str, "\r\n")

	if len(li) <= 0 {
		return errors.New("keys is null, no lines")
	}

	for _, line := range li {
		pair := strings.Split(line, ",")
		if len(pair) < 2 {
			continue
		}
		keyPair := &KeyPair{
			Id:  pair[0],
			Key: pair[1],
		}
		var err error
		keyPair.idInt, err = strconv.Atoi(keyPair.Id)
		if err != nil {
			fmt.Errorf("get invalid key %v", line)
			continue
		}
		h.keys[keyPair.Id] = keyPair
		//fmt.Printf("%v %v", pair[0], pair[1])
		h.keylist = append(h.keylist, keyPair)
	}
	if len(h.keys) <= 0 {
		return errors.New("keys is null,no keys")
	}
	h.size = len(h.keys)
	return nil
}
func (h *HadoopSconf) GetKeysSize() int {
	return len(h.keys)
}

func (h *HadoopSconf) Close() {
	if h.h != nil {
		C.destory_result(h.h)
	}
	h.h = nil
}

func (h *HadoopSconf) GetKeyPair() (pair *KeyPair) {
	//TODO  atomic
	h.pos += 1
	if h.pos > h.size {
		h.pos = 0
	}

	return h.keylist[h.pos%h.size]
}

func (h *HadoopSconf) GetKeyPairById(id string) *KeyPair {
	kp, ok := h.keys[id]
	if !ok {
		return nil
	}
	return kp
}

func (k *KeyPair) IntId() int {
	return k.idInt
}

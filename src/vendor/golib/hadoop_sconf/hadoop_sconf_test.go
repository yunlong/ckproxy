package hadoop_sconf

import (
	"fmt"
	"testing"
)

var h *HadoopSconf

func init() {
	h, _ = NewHadoopSconf("http://w-key1.safe2.shgt.qihoo.net:31500/hadoop_sconf")
}

func TestHadoopSconf(t *testing.T) {
	if h.GetKeysSize() < 10000 {
		t.Errorf("get key list failed")
	}
}

func TestHadoopSconfKeyPaire(t *testing.T) {
	for i := 0; i < 50000; i++ {
		p := h.GetKeyPair()
		if i%1000 == 0 {
			fmt.Printf("get %v %v\n", p.Id, p.Key)
		}
	}
}

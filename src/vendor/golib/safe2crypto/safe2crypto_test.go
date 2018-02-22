package safe2crypto

import (
	"bytes"
	"strings"
	"testing"
)

//func TestMakeSconfBody(t *testing.T) {
//	b := makeSconfBody()
//	s := string(b)
//	t.Logf("%s\n", s)
//	//t.Errorf("%s\n", s)
//}
//
//func TestCipherManager(t *testing.T) {
//	m, err := NewCipherManager("xxxx", "etc/security_keys.bin", "etc/client_rsa_key.bin")
//	if err != nil {
//		t.Errorf("%v %v\n", m, err)
//	}
//	err = m.getKeyList()
//	//t.Errorf("%v\n", err)
//}

func TestCipherManagerUrl(t *testing.T) {

	url := string("http://w-key1.safe2.shgt.qihoo.net:31500/hadoop_sconf")
	m, err := NewCipherManagerByUrl(url)
	if err != nil {
		t.Errorf("%v %v\n", m, err)
	}
	c, err := m.NewCipher()
	if err != nil {
		t.Errorf("%v %v\n", c, err)
	}
	src := []byte("this is a test.....")
	dst := c.Encrypt2Base64String(src)
	newSrc, err := c.DecryptBase64String(&dst)
	t.Logf("%v \n", dst)
	t.Errorf("%s\n", dst)

	t.Logf("%v %v\n", string(newSrc), err)
	if bytes.Compare(src, newSrc) != 0 {
		t.Logf("%v != %v\n", string(newSrc), string(src))
	}

	newSrc2, err := m.DecryptBase64String(&dst)
	if bytes.Compare(src, newSrc2) != 0 {
		t.Logf("%v != %v\n", string(newSrc2), string(src))
	}

	log1 := string("AAAAZAAAAUQAAABsAAAAcGWHM/IBAdb3Z3BCgKAMqKDTBPRR4mzRYFwm2LD8jMb8rXi/ywKkuIzy1NtFoCsq3zeDOabtMKemztHC3uQPNYzRMXhGz54R+lLfXlQAXIS6BvNAY9k0iPwAoXYamZOOCBP+Qg/Nj9ltjwlUpWwCCIRLisDAEcU=")
	log2 := string("AAGGqAAAA8wAAACIAAAAkBnPjokBAdoPyClBOO+I+JRWHj9+DUAUIzmmIPfUkpH7DwIyoFJF5Ig8+7ujioqC9jsT24drn47F+/l9zRgx+5pwc7XDxNpLgjBii65AoG5rhNUuJogeCNrxx0jYfAstl1J30GH9XnSUJcHYbb1zM30oD+Zb3/IOefVlUBnEURsNuMx2qyjkEoQyuHFnPD9QI6pjvjBBwQ==")
	testCiper(t, m, log1, string("abcdefadfa"))
	testCiper(t, m, log2, string("abcdefadfa"))
}

func testCiper(t *testing.T, m *CipherManager, log, compareStr string) {
	str, err := m.DecryptBase64String(&log)
	if strings.Index(string(str), compareStr) != 0 {
		t.Errorf("log is :%v\n decrylog is:%v %v\n", log, string(str), err)
	}
}

func TestInt2Byte(t *testing.T) {
	i := int32(19)
	b := int2byte4(i)
	newI := byte42int(b)
	if i != newI {
		t.Errorf("i=%v new_i=%v\n", i, newI)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	log := string("AAAAZAAAAUQAAABsAAAAcGWHM/IBAdb3Z3BCgKAMqKDTBPRR4mzRYFwm2LD8jMb8rXi/ywKkuIzy1NtFoCsq3zeDOabtMKemztHC3uQPNYzRMXhGz54R+lLfXlQAXIS6BvNAY9k0iPwAoXYamZOOCBP+Qg/Nj9ltjwlUpWwCCIRLisDAEcU=")
	for i := 0; i < b.N; i++ {
		defaultCipherManager.DecryptBase64String(&log)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	c, _ := defaultCipherManager.NewCipher()
	src := []byte("this is a test.....xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	for i := 0; i < b.N; i++ {
		c.Encrypt2Base64String(src)
	}
}

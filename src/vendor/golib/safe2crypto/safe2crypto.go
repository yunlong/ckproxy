package safe2crypto

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"golib/cgo/qhsec"
	sconf "golib/hadoop_sconf"
	"hash/crc32"
	"io"
	"net/http"
	"strings"
	"time"
)

type CipherManager struct {
	npp *qhsec.NppConfig
	h   *sconf.HadoopSconf
}

var defaultCipherManager *CipherManager = nil

func NewCipherManager(business, symmetric_key_file, asymmetric_key_file string) (m *CipherManager, err error) {
	m = &CipherManager{}

	m.npp, err = qhsec.NewNppConfig(symmetric_key_file, business, asymmetric_key_file)
	if err != nil {
		return m, err
	}
	if defaultCipherManager == nil {
		defaultCipherManager = m
	}
	return m, nil
}

func NewCipherManagerByUrl(url string) (*CipherManager, error) {
	m := &CipherManager{}
	var err error
	m.h, err = sconf.NewHadoopSconf(url)
	if defaultCipherManager == nil {
		defaultCipherManager = m
	}
	return m, err
}

func (m *CipherManager) DecryptBase64String(data *string) ([]byte, error) {
	c := &Cipher{}
	return c.DecryptBase64String(data)
}

type Cipher struct {
	k *sconf.KeyPair
	b cipher.Block
}

func (m *CipherManager) NewCipher() (*Cipher, error) {
	var err error
	c := &Cipher{}
	c.k = m.h.GetKeyPair()
	c.b, err = aes.NewCipher([]byte(c.k.Key))
	return c, err
}

// AES加密
/*
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
Key NO      Data srclen     Data compresslen    Data encryptlen     Crc32   compress type   encrypt type    encrypt Data
4 byte      4 byte          4 byte              4 byte              4 byte  1byte           1byte           -
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
compress type  0:no compress, 1:ZLIB, 2:LZO
encrypt type   0:no encrypt, 1:AES
*/

func (c *Cipher) Encrypt(src []byte) ([]byte, error) {
	key := int32(c.k.IntId())
	srcLen := int32(len(src))
	keyBuf := int2byte4(key)
	srcLenBuf := int2byte4(srcLen)

	//compress
	var compressBuf bytes.Buffer
	w := zlib.NewWriter(&compressBuf)
	w.Write(src)
	w.Close()
	compressData := compressBuf.Bytes()
	compressLen := int32(len(compressData))
	compressLenBuf := int2byte4(compressLen)

	//计算要加密的报文大小
	encryptLen := int32(0)
	if compressLen%aes.BlockSize == 0 {
		encryptLen = compressLen
	} else {
		encryptLen = (compressLen/aes.BlockSize + 1) * aes.BlockSize
		//少多少补齐
		align := make([]byte, encryptLen-compressLen)
		compressData = append(compressData, align...)
	}

	crcVal := crc32.ChecksumIEEE(src)
	crcValBuf := uint2byte4(crcVal)
	//encrypt
	encryptData, err := c.encryptIv(compressData)
	encryptType := int32(1)
	if err != nil {
		encryptData = compressData
		encryptType = 0
	}
	//加密后的报文大小
	//encryptDataLen := int32(len(encryptData))
	//fmt.Printf("======%d %d\n", encryptLen, encryptDataLen)
	encryptLenBuf := int2byte4(encryptLen)

	//1 zlib
	compressType := int32(1)
	compressTypeBuf := int2byte1(compressType)
	//1 aes
	encryptTypeBuf := int2byte1(encryptType)

	//var total []byte
	//total = append(total, keyBuf...)
	//total = append(total, srcLenBuf...)
	//total = append(total, compressLenBuf...)
	//total = append(total, encryptLenBuf...)
	//total = append(total, crcValBuf...)
	//total = append(total, compressTypeBuf...)
	//total = append(total, encryptTypeBuf...)
	//total = append(total, encryptData...)

	totalbuf := bytes.NewBuffer([]byte{})
	totalbuf.Write(keyBuf)
	totalbuf.Write(srcLenBuf)
	totalbuf.Write(compressLenBuf)
	totalbuf.Write(encryptLenBuf)
	totalbuf.Write(crcValBuf)
	totalbuf.Write(compressTypeBuf)
	totalbuf.Write(encryptTypeBuf)
	totalbuf.Write(encryptData)

	//fmt.Printf("==========%v %v %v %v %v %v %v\n", key, srcLen, compressLen, encryptLen, crcVal, compressType, encryptType)
	return totalbuf.Bytes(), nil
}

func (c *Cipher) Encrypt2Base64String(src []byte) string {
	data, err := c.Encrypt(src)
	_ = err
	str := base64.StdEncoding.EncodeToString(data)
	str += "\n"
	return str
}

//src should be base64_decode byte
func (c *Cipher) Decrypt(src []byte) ([]byte, error) {
	if len(src) < 22 {
		return []byte{}, errors.New("the data length is not enough")
	}
	keyBuf := src[0:4]
	srcLenBuf := src[4:8]
	compressLenBuf := src[8:12]
	encryptLenBuf := src[12:16]
	crcValBuf := src[16:20]
	compressTypeBuf := src[20:21]
	encryptTypeBuf := src[21:22]
	encryptData := src[22:]
	_ = encryptData

	key := byte42int(keyBuf)
	keyStr := fmt.Sprintf("%d", key)
	srcLen := byte42int(srcLenBuf)
	compressLen := byte42int(compressLenBuf)
	encryptLen := byte42int(encryptLenBuf)
	crcVal := byte42uint(crcValBuf)
	compressType := byte2int(compressTypeBuf)
	encryptType := byte2int(encryptTypeBuf)

	if c.k == nil {
		c.k = defaultCipherManager.h.GetKeyPairById(keyStr)
		if c.k == nil {
			return []byte{}, fmt.Errorf("the key %d is not found", key)
		}
		var err error
		c.b, err = aes.NewCipher([]byte(c.k.Key))
		//fmt.Printf("keyis %v %v\n", c.k.Key, len([]byte(c.k.Key)))
		if err != nil {
			return []byte{}, fmt.Errorf("the key %d get failed:%v", err)
		}
	}

	//fmt.Printf("=decrpyt=========%v %v %v %v %v %v %v %v %v\n", key, srcLen, compressLen, encryptLen, crcVal, compressType, encryptType, len(src), aes.BlockSize)

	_ = compressLen
	_ = encryptLen
	_ = crcVal

	var decryptData []byte
	var plainData []byte
	var err error
	if encryptType == 1 {
		decryptData, err = c.decryptIv(encryptData)
		if err != nil {
			return plainData[:srcLen], err
		}

		//fmt.Printf("decrpyt= %v %v\n", decryptData, len(decryptData))
	}

	if compressType == 1 {
		b := bytes.NewReader(decryptData[:compressLen])
		r, err := zlib.NewReader(b)
		defer r.Close()
		if err != nil {
			plainData = decryptData
			return plainData, err
		} else {
			plainData = make([]byte, srcLen)
			_, err := io.ReadFull(r, plainData)
			if err != nil {
				return plainData, err
			}
		}
	} else {
		plainData = decryptData
	}
	return plainData, nil
}

func (c *Cipher) DecryptBase64String(src *string) ([]byte, error) {
	str := strings.Trim(*src, "\n")
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return data, err
	}
	return c.Decrypt(data)
}

func int2byte4(val int32) []byte {
	buf := make([]byte, 4)
	buf[3] = byte(val >> 0 & 0xff)
	buf[2] = byte(val >> 8 & 0xff)
	buf[1] = byte(val >> 16 & 0xff)
	buf[0] = byte(val >> 24 & 0xff)
	return buf
}

func uint2byte4(val uint32) []byte {
	buf := make([]byte, 4)
	buf[3] = byte(val >> 0 & 0xff)
	buf[2] = byte(val >> 8 & 0xff)
	buf[1] = byte(val >> 16 & 0xff)
	buf[0] = byte(val >> 24 & 0xff)
	return buf
}

func int2byte1(val int32) []byte {
	buf := make([]byte, 1)
	buf[0] = byte(val >> 0 & 0xff)
	return buf
}

func byte42int(val []byte) int32 {
	return (int32(val[0]) << 24) | (int32(val[1]) << 16) | (int32(val[2]) << 8) | (int32(val[3]) << 0)
}

func byte42uint(val []byte) uint32 {
	return (uint32(val[0])&0xff)<<24 |
		(uint32(val[1])&0xff)<<16 |
		(uint32(val[2])&0xff)<<8 |
		(uint32(val[3])&0xff)<<0
}

func byte2int(val []byte) int {
	return int(val[0])
}

//采用固定iv
func (c *Cipher) encryptIv(src []byte) ([]byte, error) {
	// 验证输入参数
	// 必须为aes.BlockSize的倍数
	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	encryptText := make([]byte, aes.BlockSize+len(src))

	iv := make([]byte, aes.BlockSize)
	for i := 0; i < aes.BlockSize; i++ {
		iv[i] = 0
	}

	mode := cipher.NewCBCEncrypter(c.b, iv)

	mode.CryptBlocks(encryptText, src)

	return encryptText, nil
}

//采用固定iv
func (c *Cipher) decryptIv(src []byte) (dst []byte, err error) {
	ciphertext := src
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	totalLen := len(ciphertext)
	if totalLen < aes.BlockSize {
		return dst, errors.New("ciphertext too short")
	}
	iv := make([]byte, aes.BlockSize)
	for i := 0; i < aes.BlockSize; i++ {
		iv[i] = 0
	}

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		return dst, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(c.b, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)
	dst = ciphertext
	return dst, err
}

func (c *Cipher) encrypt(src []byte) ([]byte, error) {
	// 验证输入参数
	// 必须为aes.BlockSize的倍数
	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	encryptText := make([]byte, aes.BlockSize+len(src))

	iv := encryptText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(c.b, iv)

	mode.CryptBlocks(encryptText[aes.BlockSize:], src)

	return encryptText, nil
}

func (c *Cipher) decrypt(src []byte) (dst []byte, err error) {
	ciphertext := src
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	totalLen := len(ciphertext)
	if totalLen < aes.BlockSize {
		return dst, errors.New("ciphertext too short")
	}
	//iv := ciphertext[:aes.BlockSize]
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(ciphertext)%aes.BlockSize != 0 {
		return dst, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(c.b, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, ciphertext)
	dst = ciphertext
	return dst, err
}

func makeSconfBody() []byte {
	body := []byte{}
	t := fmt.Sprintf("%d", time.Now().Unix())
	vk := []byte{}
	vk = append(vk, []byte(t)[:6]...)
	vk = append(vk, []byte("hadoop_security")...)
	vkMd5 := fmt.Sprintf("%x", md5.Sum(vk))
	body = append(body, []byte("product=hadoop\r\ntime=")...)
	body = append(body, t...)
	body = append(body, []byte("\r\nvk=")...)
	body = append(body, vkMd5[:]...)
	body = append(body, []byte("\r\n")...)

	return body
}

func (m *CipherManager) makeSconfCryptoBody() ([]byte, error) {
	body := makeSconfBody()
	cli, err := qhsec.NewClientPacker(m.npp, 6)
	defer cli.Close()
	if err != nil {
		return []byte{}, err
	}
	err = cli.SetOption(qhsec.OptAsymmetricMethod, qhsec.NaclEC)
	fmt.Printf("%v\n", err)
	err = cli.SetOption(qhsec.OptSymmetricMethod, qhsec.IDEAECB) // idea
	fmt.Printf("%v\n", err)
	err = cli.SetOption(qhsec.OptSymmetricKeyNo, 1)
	fmt.Printf("%v\n", err)
	err = cli.SetOption(qhsec.OptAsymmetricKeyNo, 1)
	fmt.Printf("%v\n", err)
	err = cli.SetOption(qhsec.OptNetMethod, 5)
	fmt.Printf("%v\n", err)
	fmt.Printf("%v\n%v %v\n", err, body, len(body))
	cryptBody, err := cli.Pack(body)
	fmt.Printf("%v\n%v\n", err, cryptBody)
	return body, err
}

func (m *CipherManager) getKeyList() error {
	b, err := m.makeSconfCryptoBody()
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(b)
	resp, err := http.Post("http://w-key1.safe2.shgt.qihoo.net:31500/hadoop_sconf", "", buffer)
	_ = resp
	//fmt.Printf("%v %v\n", resp, err)
	return err
}

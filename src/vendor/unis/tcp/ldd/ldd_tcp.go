package ldd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	_ "hash/crc32"
	"io"
	"log"
	_ "time"
	"unis/tcp"
)

var _ = log.Print
var emptyExtras = make(tcp.Extras)

type Header struct { // See LDD TCP protocol
	Version     uint8
	Type        uint8
	sequence    uint32
	payloadType uint16
	payloadSize uint32
	checksum    uint16 // crc16
}

type Body struct {
	payload []byte
}

type LddHeader struct { // See LDD TCP protocol
	Header
	ExtraN    uint8
	ExtraSize uint16
}

type LddTCP struct {
	LddHeader
	Body
	recvExtras tcp.Extras // The received extra data
	sendExtras tcp.Extras // The sent extra data
}

const (
	headerLen = 16

	// 请求方向
	kRequest = 0
	kPing    = 1
	kCancel  = 2

	// 响应方向
	kResponse = 128
	kPong     = 129
	kLast     = 130
	kEnd      = 131
)

func NewLddTCP() tcp.TCPCodec {
	return &LddTCP{
		recvExtras: make(tcp.Extras, 3),
		sendExtras: make(tcp.Extras, 3),
	}
}

func MakeLddTCP() *LddTCP {
	return &LddTCP{
		recvExtras: make(tcp.Extras, 3),
		sendExtras: make(tcp.Extras, 3),
	}
}

func (t *LddTCP) ReadHeaderFromBuf(buf []byte) (err error) {
	t.Version = uint8(buf[0])
	if t.Version != 0xA1 {
		return fmt.Errorf("ERROR version %v", t.Version)
	}
	t.Type = uint8(buf[1])
	t.sequence = uint32(buf[2])<<24 | uint32(buf[3])<<16 | uint32(buf[4])<<8 | uint32(buf[5])
	t.payloadType = uint16(buf[6])<<8 | uint16(buf[7])
	t.payloadSize = uint32(buf[8])<<16 | uint32(buf[9])<<8 | uint32(buf[10])
	t.checksum = uint16(buf[14])<<8 | uint16(buf[15])
	t.ExtraN = uint8(buf[11])
	t.ExtraSize = uint16(buf[12])<<8 | uint16(buf[13])

	// Check the crc16
	c16 := ChecksumCCITT(buf[:14])
	if uint16(c16) != t.checksum {
		return fmt.Errorf("Checksum ERROR. recv %v, calculate %v", t.checksum, c16)
	}
	fmt.Printf("ParseHeaderFromBuf success LddTcp:%v\n", t)
	fmt.Printf("t.payloadType:%v\n", t.payloadType)
	return nil
}

func (t *LddTCP) ReadBodyFromBuf(buf []byte) (err error) {
	// Read all body data

	// Parse payload
	t.payload = buf[headerLen : t.payloadSize+headerLen]

	buf2 := buf[headerLen:]

	// Read extras
	ibeg := t.payloadSize
	ebeg := t.payloadSize + uint32(t.ExtraN)*2
	for i := uint8(0); i < t.ExtraN; i++ {
		id := buf2[ibeg]
		exlen := uint32(buf2[ibeg+1])
		ibeg += 2

		t.recvExtras[int(id)] = buf2[ebeg:(ebeg + exlen)]
		ebeg += exlen
	}
	//fmt.Printf("ParseBodyFromBuf success LddTcp:%v\n", t);

	return nil
}
func (t *LddTCP) ReadHeader(r io.Reader) (err error) {
	buf := make([]byte, headerLen)
	n, err := io.ReadFull(r, buf)
	if err != nil || n != headerLen {
		if err == io.EOF {
			return err
		}
		return fmt.Errorf("LDD TCP header read %v bytes ERROR failed : %v", headerLen, err.Error())
	}

	// decode header
	t.Version = uint8(buf[0])
	if t.Version != 0xA1 {
		return fmt.Errorf("ERROR version %v", t.Version)
	}
	t.Type = uint8(buf[1])
	t.sequence = uint32(buf[2])<<24 | uint32(buf[3])<<16 | uint32(buf[4])<<8 | uint32(buf[5])
	t.payloadType = uint16(buf[6])<<8 | uint16(buf[7])
	t.payloadSize = uint32(buf[8])<<16 | uint32(buf[9])<<8 | uint32(buf[10])
	t.checksum = uint16(buf[14])<<8 | uint16(buf[15])
	t.ExtraN = uint8(buf[11])
	t.ExtraSize = uint16(buf[12])<<8 | uint16(buf[13])

	// Check the crc16
	c16 := ChecksumCCITT(buf[:14])
	if uint16(c16) != t.checksum {
		return fmt.Errorf("Checksum ERROR. recv %v, calculate %v", t.checksum, c16)
	}
	return nil
}

func (t *LddTCP) ReadBody(r io.Reader) (err error) {
	// Read all body data
	len := int(t.payloadSize + uint32(t.ExtraN)*2 + uint32(t.ExtraSize))
	buf := make([]byte, len)
	n, err := io.ReadFull(r, buf)
	if err != nil || n != len {
		return fmt.Errorf("LDD TCP body read %v bytes ERROR failed : %v", t.payloadSize, err.Error())
	}

	// Parse payload
	t.payload = buf[:t.payloadSize]

	// Read extras
	ibeg := t.payloadSize
	ebeg := t.payloadSize + uint32(t.ExtraN)*2
	for i := uint8(0); i < t.ExtraN; i++ {
		id := buf[ibeg]
		exlen := uint32(buf[ibeg+1])
		ibeg += 2

		t.recvExtras[int(id)] = buf[ebeg:(ebeg + exlen)]
		ebeg += exlen
	}

	return nil
}

func (t *LddTCP) Handle(h tcp.CallbackHandler) (send []byte, err error) {
	if t.Type == kPing {
		//TODO delte this debug log
		//fmt.Errorf("================================ping ...")
		return nil, nil
	}

	if t.Type == kCancel {
		//no need to process this message now
		return nil, nil
	}

	return h.CallbackHandle()
}

func (t *LddTCP) PayloadType() uint16 {
	fmt.Printf("PayloadType::payloadType:%v\n", t.payloadType)
	return t.payloadType
}

func (t *LddTCP) Payload() []byte {
	return t.payload
}

func (t *LddTCP) GetRecvExtras() tcp.Extras {
	return t.recvExtras
}

func (t *LddTCP) AddSendExtra(id int, d []byte) {
	t.sendExtras[id] = d
}

func (t *LddTCP) SetSendExtra(extras tcp.Extras) {
	if extras == nil || len(extras) == 0 {
		return
	}
	t.sendExtras = extras
}

func (t *LddTCP) ClearSendExtra() {
	if len(t.sendExtras) == 0 {
		return
	}
	t.sendExtras = make(tcp.Extras, 3)
}

func WriteN(w io.Writer, buff []byte) (n int, err error) {
	//return w.Write(buff)
	size := len(buff)
	if size == 0 {
		return 0, nil
	}
	writen := 0
	n = 0
	for writen < size {
		n, err = w.Write(buff[writen:])
		if err != nil {
			fmt.Printf("=============>total=%v writen=%v curn=%v err=%v \n", size, writen, n, err)
			//TODO EGAIN
			break
		}
		writen += n
	}

	return writen, err
}

func (t *LddTCP) Write(payload []byte, w io.Writer) error {
	if t.Type == kResponse || t.Type == kLast || t.Type == kEnd {
		return nil
	}

	err := t.writeHeader(w, t.sequence, uint32(len(payload)), t.sendExtras)
	if err != nil {
		return err
	}
	if payload != nil {
		_, err := WriteN(w, payload)
		if err != nil {
			return err
		}
	}
	return t.writeExtra(w)
}

//request form client to  server:  incoming, c -> s;
//request from server to  client:  outcoming, s -> c;
//c -> s 方向上， 发送回应数据;
//在 s->c 方向上，发送请求数据;
//c->s s->c 两个方向上的sequence，不一定相等，单方向上的sequence，要递增;
func (t *LddTCP) WriteEx(payload []byte, w io.Writer, payloadType int, request bool) error {
	err := t.writeHeaderEx(w, t.sequence, uint32(len(payload)), t.sendExtras, payloadType, request)
	if err != nil {
		return err
	}
	if payload != nil {
		_, err := WriteN(w, payload)
		if err != nil {
			return err
		}
	}
	return t.writeExtra(w)
}

//表示在c -> s 方向上， 发送回应数据
func (t *LddTCP) writeHeader(w io.Writer, sequence uint32, payloadSize uint32, extra tcp.Extras) error {
	return t.writeHeaderEx(w, sequence, payloadSize, extra, 0, false)
}

func (t *LddTCP) writeEnd(w io.Writer, sequence uint32) error {
	return t.writeHeader(w, sequence, 0, emptyExtras)
}

func int82Byte(i uint8) byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, i)
	return buf.Bytes()[0]
}

func (t *LddTCP) writeExtra(w io.Writer) error {
	if len(t.sendExtras) == 0 {
		return nil
	}

	// write Extras-Index
	var dataBuf []byte
	buf := make([]byte, 2*len(t.sendExtras))
	id := int(0)
	for k, v := range t.sendExtras {
		buf[id*2] = int82Byte(uint8(k))
		buf[id*2+1] = int82Byte(uint8(len(v)))
		id += 1
		dataBuf = append(dataBuf, v[:]...)
	}

	_, err := WriteN(w, buf)
	if err != nil {
		return err
	}

	_, err = WriteN(w, dataBuf)
	if err != nil {
		return err
	}
	return nil
}

func fillChecksum(buf []byte) {
	c16 := ChecksumCCITT(buf[:14])
	buf[14] = byte(c16 >> 8)
	buf[15] = byte(c16 & 0xFF)
}

//for outcoming request;
func (t *LddTCP) IncraseSeq() {
	//todo must be automic
	t.sequence++
}

//request = true 表示发请求数据
func (t *LddTCP) writeHeaderEx(w io.Writer, sequence uint32, payloadSize uint32, extra tcp.Extras, payloadType int, request bool) error {
	buf := make([]byte, headerLen)
	buf[0] = byte(t.Version)

	if !request {
		if t.Type == kPing {
			buf[1] = kPong
		} else if payloadSize == 0 && t.Type == kRequest {
			buf[1] = kEnd
		} else {
			buf[1] = kLast
		}
	} else {
		buf[1] = t.Type
	}

	buf[2] = byte(sequence >> 24)
	buf[3] = byte(sequence >> 16 & 0xFF)
	buf[4] = byte(sequence >> 8 & 0xFF)
	buf[5] = byte(sequence & 0xFF)
	if payloadType != 0 {
		var payloadType16 int16 = int16(payloadType)
		b_buf := bytes.NewBuffer([]byte{})
		binary.Write(b_buf, binary.BigEndian, payloadType16)
		buf[6] = b_buf.Bytes()[0]
		buf[7] = b_buf.Bytes()[1]
	} else if payloadSize == 0 && t.Type == kRequest {
		//err code = 1
		var err16 int16 = int16(1)
		b_buf := bytes.NewBuffer([]byte{})
		binary.Write(b_buf, binary.BigEndian, err16)
		buf[6] = b_buf.Bytes()[0]
		buf[7] = b_buf.Bytes()[1]
	} else {
		//Payload-Type must be 0
		buf[6] = byte(0)
		buf[7] = byte(0)
	}
	buf[8] = byte(payloadSize >> 16 & 0xFF)
	buf[9] = byte(payloadSize >> 8 & 0xFF)
	buf[10] = byte(payloadSize & 0xFF)
	buf[11] = byte(len(extra))

	extraSize := 0
	for _, v := range extra {
		extraSize += len(v)
	}

	buf[12] = byte(extraSize >> 8)
	buf[13] = byte(extraSize & 0xFF)

	//var extraSize16 int16 = int16(extraSize)
	//b_buf := bytes.NewBuffer([]byte{})
	//binary.Write(b_buf, binary.BigEndian, extraSize16)
	//fmt.Printf("-========================> %v\n", b_buf.Bytes())
	//fmt.Printf("-========================> %v %v\n", buf[12], buf[13])
	fillChecksum(buf)
	WriteN(w, buf)
	return nil
}

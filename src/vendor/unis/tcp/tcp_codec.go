package tcp

import (
	"io"
)

type TCPAccessor interface {
	Payload() []byte
	PayloadType() uint16           // The payload type which can be used as a network method
	GetRecvExtras() Extras         // The receiving extra data
	AddSendExtra(id int, d []byte) // Add an extra to send
}

type TCPCodec interface {
	// Read
	ReadHeader(r io.Reader) error
	ReadBody(r io.Reader) error

	Handle(h CallbackHandler) (send []byte, err error)

	// Accessor for unis.Module
	TCPAccessor

	// Write
	Write(payload []byte, w io.Writer) error
}

type TCPCodecFactory func() TCPCodec

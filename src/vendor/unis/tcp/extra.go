package tcp

type Extras map[int][]byte

type CallbackHandler interface {
	CallbackHandle() (send []byte, err error)
}

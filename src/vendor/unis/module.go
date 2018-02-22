package unis

import (
//	"golib/cgo/qhsec"
)

type Module interface {
	Initialize() error
//	NppConfig() *qhsec.NppConfig
}

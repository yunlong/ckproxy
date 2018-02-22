package stat

import (
	"fmt"
	"runtime"
	"sync"
)

type FuncCallStat struct {
	statsIn  map[string]uint64
	statsOut map[string]uint64
	lockIn   sync.Mutex
	lockOut  sync.Mutex
	Enable   bool
}

func NewFuncCallStat() *FuncCallStat {
	return &FuncCallStat{
		statsIn:  map[string]uint64{},
		statsOut: map[string]uint64{},
		Enable:   true,
	}
}
func (s *FuncCallStat) Enter() {
	if !s.Enable {
		return
	}
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	s.lockIn.Lock()
	defer s.lockIn.Unlock()
	s.statsIn[f.Name()] += 1
}

func (s *FuncCallStat) Leave() {
	if !s.Enable {
		return
	}
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	s.lockOut.Lock()
	defer s.lockOut.Unlock()
	s.statsOut[f.Name()] += 1
}

func (s *FuncCallStat) Dump() string {
	info := string("")
	for key, count := range s.statsIn {
		info += fmt.Sprintf("%v enter %v\n", key, count)
	}

	for key, count := range s.statsOut {
		info += fmt.Sprintf("%v leave %v\n", key, count)
	}
	return info
}

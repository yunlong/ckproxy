package stat

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// // Example:
//	t := NewTimer()
//	t.Add("total")
//	var wg sync.WaitGroup
//	for _, name := range []string{"first", "second", "three", "four"} {
//		wg.Add(1)
//		go func(n string) {
//			defer wg.Done()
//			t.Add(n)
//			defer t.Stop(n)
//			time.Sleep(time.Duration(len(n)) * time.Millisecond)
//		}(name)
//	}
//
//	wg.Wait()
//	t.StopAll()
//	fmt.Println(t.Dump())
//  // Output:
//  // first=5178435    four=4182445    second=6117970  three=5108715   total=6317020
type Timer struct {
	lock sync.Mutex
	pool map[string]*timeRange
}

type timeRange struct {
	begin    time.Time
	duration time.Duration
}

func NewTimer() *Timer {
	return &Timer{
		pool: make(map[string]*timeRange, 0),
	}
}

func (t *Timer) Add(name string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.pool[name] = &timeRange{
		begin:    time.Now(),
		duration: -1,
	}
}

func (t *Timer) Stop(name string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if k, ok := t.pool[name]; ok {
		k.duration = time.Since(k.begin)
	}
}

func (t *Timer) StopAll() {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, v := range t.pool {
		if v.duration != -1 {
			continue
		}
		v.duration = time.Since(v.begin)
	}
}

func (t *Timer) Dump() string {
	t.lock.Lock()
	defer t.lock.Unlock()
	buf := make([]string, 0)
	for k, v := range t.pool {
		buf = append(buf, fmt.Sprintf("%v=%v",
			k, v.duration.Nanoseconds()))
	}

	sort.Strings(buf)
	return strings.Join(buf, "\t")
}

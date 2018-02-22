package stat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type StatHelper struct {
	countRecordsMutex sync.RWMutex
	countRecords      map[string]*countRecord

	timeRecordsMutex sync.RWMutex
	timeRecords      map[string]*timeRecord

	timer *time.Timer
}

const (
	timeCost = 0
	count    = 1
)

type statInfo struct {
	statType int
	name     string
	duration time.Duration
}

type countRecord struct {
	Count int64
}

type timeRecord struct {
	Times         int64 //count
	TotalUsedTime int64
	MaxUsedTime   int64
	MinUsedTime   int64
}

func NewStatHelper() (st *StatHelper) {
	st = new(StatHelper)
	st.countRecords = make(map[string]*countRecord)
	st.timeRecords = make(map[string]*timeRecord)
	return
}

func (sh *StatHelper) Close() {
	if sh.timer != nil {
		sh.timer.Stop()
	}
}

func (sh *StatHelper) SetTimerDump(duration time.Duration, f func()) {
	sh.timer = time.AfterFunc(duration, func() {
		f()
		sh.countRecordsMutex.Lock()
		sh.countRecords = make(map[string]*countRecord)
		sh.countRecordsMutex.Unlock()

		sh.timeRecordsMutex.Lock()
		sh.timeRecords = make(map[string]*timeRecord)
		sh.timeRecordsMutex.Unlock()
		sh.timer.Reset(duration)
	})
}

func (sh *StatHelper) AddCount(name string) {
	sh.AddCountN(name, 1)
}

func (sh *StatHelper) AddCountN(name string, count int) {
	cnt := sh.getCountRecord(name)
	atomic.AddInt64(&cnt.Count, int64(count))
}

func (sh *StatHelper) addTimeCostCount(name string, usedNano int64) {
	sh.AddCount(name)
	switch {
	case usedNano <= int64(10*1000*1000):
		//[0 ms,10 ms]
		sh.AddCount(name + "_10")
	case usedNano > int64(10*1000*1000) && usedNano <= int64(100*1000*1000):
		//(10 ms,100 ms]
		sh.AddCount(name + "_100")
	case usedNano > int64(100*1000*1000) && usedNano <= int64(200*1000*1000):
		//(100 ms,200 ms]
		sh.AddCount(name + "_200")
	case usedNano > int64(200*1000*1000) && usedNano <= int64(300*1000*1000):
		//(200 ms,300 ms]
		sh.AddCount(name + "_300")
	case usedNano > int64(300*1000*1000) && usedNano <= int64(400*1000*1000):
		//(300 ms,400 ms]
		sh.AddCount(name + "_400")
	case usedNano > int64(400*1000*1000) && usedNano <= int64(500*1000*1000):
		//(400 ms,500 ms]
		sh.AddCount(name + "_500")
	case usedNano > int64(500*1000*1000) && usedNano <= int64(1000*1000*1000):
		// (500 ms,1s]
		sh.AddCount(name + "_1000")
	case usedNano > int64(1000*1000*1000):
		//more than 1s
		sh.AddCount(name + "_M1000")
	}
}

func (sh *StatHelper) AddTimeStat(name string, usedTime time.Duration) {
	r := sh.getTimeRecord(name)
	usedNano := usedTime.Nanoseconds()
	sh.addTimeCostCount(name, usedNano)
	atomic.AddInt64(&r.Times, 1)
	atomic.AddInt64(&r.TotalUsedTime, usedNano)
	for {
		old := atomic.LoadInt64(&r.MaxUsedTime)
		if old >= usedNano || atomic.CompareAndSwapInt64(&r.MaxUsedTime, old, usedNano) {
			break
		}
	}
	for {
		old := atomic.LoadInt64(&r.MinUsedTime)
		if old <= usedNano || atomic.CompareAndSwapInt64(&r.MinUsedTime, old, usedNano) {
			break
		}
	}
}

func (sh *StatHelper) DumpCount(writer io.Writer) error {
	buffio := bufio.NewWriter(writer)
	sh.countRecordsMutex.RLock()
	defer sh.countRecordsMutex.RUnlock()
	if len(sh.countRecords) == 0 {
		return errors.New("stat empty")
	}

	for name, count := range sh.countRecords {
		if _, err := fmt.Fprintf(buffio, "%s_c=%d\t", name, count.Count); err != nil {
			return err
		}
	}
	return buffio.Flush()
}

func (sh *StatHelper) DumpTimeCost(writer io.Writer) error {
	results := sh.getTimeRecords()
	sort.Sort(results)
	buf := bufio.NewWriter(writer)
	//if _, err := fmt.Fprintln(writer, "name,times,avg,min,max,total"); err != nil {
	//	return err
	//}
	for _, r := range results {
		if _, err := fmt.Fprintf(writer,
			"%s=%d,%.2f,%.3f,%.2f\t",
			r.Name,
			r.Times,
			float64(r.AvgUsedTime)/1000000,
			float64(r.MinUsedTime)/1000000,
			float64(r.MaxUsedTime)/1000000,
			//float64(r.TotalUsedTime)/1000000,
		); err != nil {
			return err
		}
	}
	return buf.Flush()
}

func (sh *StatHelper) DumpSimpleTimeCost(writer io.Writer) error {
	totalCount := int64(0)
	results := sh.getTimeRecords()
	sort.Sort(results)
	buf := bufio.NewWriter(writer)
	//if _, err := fmt.Fprintln(writer, "name,times,avg,min,max,total"); err != nil {
	//	return err
	//}
	if len(results) == 0 {
		return errors.New("stat empty")
	}
	for _, r := range results {
		if r.Times > totalCount {
			totalCount = r.Times
		}
		if _, err := fmt.Fprintf(writer,
			"%s=%f\t",
			r.Name,
			float64(r.AvgUsedTime)/1000000,
		); err != nil {
			return err
		}
	}
	fmt.Fprintf(writer, "total_count=%d\t", totalCount)
	return buf.Flush()
}

func (sh *StatHelper) getCountRecord(name string) *countRecord {
	sh.countRecordsMutex.Lock()
	defer sh.countRecordsMutex.Unlock()
	r, exists := sh.countRecords[name]
	if !exists {
		r = new(countRecord)
		sh.countRecords[name] = r
	}
	return r
}

func (sh *StatHelper) getTimeRecord(name string) *timeRecord {
	sh.timeRecordsMutex.Lock()
	defer sh.timeRecordsMutex.Unlock()
	r, exists := sh.timeRecords[name]
	if !exists {
		r = new(timeRecord)
		sh.timeRecords[name] = r
	}
	return r
}

func (sh *StatHelper) getTimeRecords() sortTimeRecords {
	sh.timeRecordsMutex.RLock()
	defer sh.timeRecordsMutex.RUnlock()
	results := make(sortTimeRecords, 0, len(sh.timeRecords))
	for name, d := range sh.timeRecords {
		results = append(results, &sortTimeRecord{
			Name:          name,
			Times:         d.Times,
			AvgUsedTime:   float64(d.TotalUsedTime) / float64(d.Times),
			MaxUsedTime:   d.MaxUsedTime,
			MinUsedTime:   d.MinUsedTime,
			TotalUsedTime: d.TotalUsedTime,
		})
	}
	return results
}

type sortTimeRecord struct {
	Name          string
	Times         int64 //count
	AvgUsedTime   float64
	MinUsedTime   int64
	MaxUsedTime   int64
	TotalUsedTime int64
}

type sortTimeRecords []*sortTimeRecord

func (this sortTimeRecords) Len() int {
	return len(this)
}
func (this sortTimeRecords) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
func (this sortTimeRecords) Less(i, j int) bool {
	return this[i].AvgUsedTime > this[j].AvgUsedTime || (this[i].AvgUsedTime == this[j].AvgUsedTime && this[i].Times < this[j].Times)
}

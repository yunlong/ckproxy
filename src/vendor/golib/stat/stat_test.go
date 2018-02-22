package stat

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCount(t *testing.T) {
	var wg sync.WaitGroup
	sh := NewStatHelper()
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < i; n++ {
				sh.AddCountN("test", n)
				sh.AddCountN("test2", n)
			}
		}()
	}
	wg.Wait()
	w := bufio.NewWriter(os.Stdout)

	sh.DumpCount(w)
	fmt.Fprintf(w, "\n")
	w.Flush()
}

func TestTimeRecorder(t *testing.T) {
	sh := NewStatHelper()

	sh.SetTimerDump(time.Second, func() {
		w := bufio.NewWriter(os.Stdout)
		fmt.Fprintf(w, "count dump:\n")
		sh.DumpCount(w)
		fmt.Fprintf(w, "\ntime cost dump\nname=times,avg(ms),min(ms),max(ms)\n")
		sh.DumpTimeCost(w)
		fmt.Fprintf(w, "\n")
		w.Flush()
	})

	for n := 0; n < 2; n++ {
		for i := 0; i < 10; i++ {
			t := time.Now()
			doSomething(100)
			sh.AddTimeStat("doSomething(100)", time.Since(t))
		}
		for i := 0; i < 10; i++ {
			t := time.Now()
			doSomething(1000)
			sh.AddTimeStat("doSomething(1000)", time.Since(t))
		}
		for i := 0; i < 10; i++ {
			t := time.Now()
			doSomething(10000)
			sh.AddTimeStat("doSomething(10000)", time.Since(t))
		}
		for i := 0; i < 10; i++ {
			t := time.Now()
			doSomething(100000)
			sh.AddTimeStat("doSomething(100000)", time.Since(t))
		}

		//w := bufio.NewWriter(os.Stdout)
		w := bytes.NewBuffer([]byte{})
		sh.DumpCount(w)
		fmt.Fprintf(w, "\n")
		sh.DumpTimeCost(w)
		//w.Flush()
		fmt.Printf("=====================\n%v===========\n", string(w.Bytes()))
		time.Sleep(time.Second)
	}
}
func doSomething(n int) int {
	m := 0
	for i := 0; i < n; i++ {
		m += i
		for j := 0; j < 30; j++ {
			strconv.Itoa(m) // cost some CPU time
		}
	}
	return m
}

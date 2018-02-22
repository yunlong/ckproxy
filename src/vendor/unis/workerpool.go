/**
 * author:yunlong.Lee <yunlong.lee@163.com>
 */

package unis

import (
	"sync"
	"time"
    "github.com/golang/glog"
)

type WorkerPool struct {

	max_worker_count        int            //the max worker number in the pool
	cur_worker_count        int            //the worker number in the pool
    max_worker_idle_time    time.Duration

	lock                    sync.Mutex     //for queue thread-safe
	readyQue                []*Worker      //the available worker queue

    stopCh                  chan struct{}
    mustStop                bool

	workerPool              sync.Pool

}

type Worker struct {
	fn              chan func()
	last_used_time  int64
}

func NewWorkerPoolLimit(maxWorkerNum, maxWorkerIdleTime int) (*WorkerPool, error) {

	wp := &WorkerPool{
		max_worker_count: maxWorkerNum,
        max_worker_idle_time: time.Duration(maxWorkerIdleTime) * time.Second,
	}
	wp.init()

	return wp, nil
}

func (wp *WorkerPool) init() {

    if wp.stopCh != nil {
        panic("BUG: WorkerPool already started")
    }

    wp.stopCh = make(chan struct{})
    stopCh := wp.stopCh
    go func() {
        var scratch []*Worker
        for {
            wp.clean(&scratch)
            select {
            case <-stopCh:
                return
            default:
                time.Sleep( wp.get_max_worker_idle_time() )
            }
        }
    }()

}

func (wp *WorkerPool) get_max_worker_idle_time() time.Duration {
    if wp.max_worker_idle_time <= 0 {
            return 10 * time.Second
    }
    return wp.max_worker_idle_time
}

func (wp *WorkerPool) Stop() {

    if wp.stopCh == nil {
        panic("BUG: workerPool wasn't started")
    }
    close(wp.stopCh)
    wp.stopCh = nil

    // Stop all the workers waiting for incoming connections.
    // Do not wait for busy workers - they will stop after
    // serving the connection and noticing wp.mustStop = true.
    wp.lock.Lock()
    readyQ := wp.readyQue
    for i, w := range readyQ {
        w.fn <- nil
        readyQ[i] = nil
    }
    wp.readyQue = readyQ[:0]
    wp.mustStop = true
    wp.lock.Unlock()

}

/**
 * cleaning up goroutine for worker
 */
func (wp *WorkerPool) clean(scratch *[]*Worker) {

    glog.Infof("starting clean up workers")

    max_worker_idle_time := wp.get_max_worker_idle_time()

    // clean least recently used workers if they didn't serve connections
    // for more than max_worker_idle_time
    current_time := time.Now().Unix()
    ////////////////////////////////////////////////////////////////
    wp.lock.Lock()
    readyQ := wp.readyQue
    n := len(readyQ)
    i := 0

    for i < n && time.Duration(current_time - readyQ[i].last_used_time) * time.Second > max_worker_idle_time {
        i++
    }

    *scratch = append((*scratch)[:0], readyQ[:i]...)
    if i > 0 {
        m := copy(readyQ, readyQ[i:])
        for i = m; i < n; i++ {
            readyQ[i] = nil
        }
        wp.readyQue = readyQ[:m]
    }
    wp.lock.Unlock()
    //////////////////////////////////////////////////////////////////
    // Notify obsolete workers to stop.
    // This notification must be outside the wp.lock, since Worker.fn
    // may be blocking and may consume a lot of time if many workers
    // are located on non-local CPUs.
    //////////////////////////////////////////////////////////////////
    tmp := *scratch
    for i, w := range tmp {
        w.fn <- nil
        tmp[i] = nil
    }
    //////////////////////////////////////////////////////////////////

    glog.Infof("worker pool : max_worker_idle_time %s; max_worker_count=%d, cur_worker_count=%d",
                              wp.max_worker_idle_time, wp.max_worker_count, len(wp.readyQue) )
}

// Serve assigns a worker for job (fn func(), with closure we can define every job in this form)
// If the worker pool is limited-number and the worker number has reached the limit, we prefer to discard the job.
func (wp *WorkerPool) Serve(fn func()) bool {

    worker := wp.getWorker()
    if worker == nil {
        return false
    }
    worker.fn <-fn

    return true
}

/**
 * 1. getWorker select a worker.
 * 2. getWorker starts a new goroutine.
 * 3. ReadyQue is like a FILO queue, and the select algorithm is kind of like LRU.
 */
func (wp *WorkerPool) getWorker() *Worker {

    var wch *Worker
    createWorker := false

    wp.lock.Lock()
    readyQ := wp.readyQue
    n := len(readyQ) - 1
    if n < 0 {
        if wp.cur_worker_count < wp.max_worker_count {
            createWorker = true
            wp.cur_worker_count++
        }
    } else {
        wch = readyQ[n]
        readyQ[n] = nil
        wp.readyQue = readyQ[:n]
    }
    wp.lock.Unlock()

    if wch == nil {
        if !createWorker {
            return nil
        }

        vch := wp.workerPool.Get()
        if vch == nil {
            vch = &Worker{
                fn: make(chan func()),
            }
        }

        wch = vch.(*Worker)

        go func() {
            wp.workerFunc(wch)
            wp.workerPool.Put(vch)
        }()
    }

    return wch
}

func (wp *WorkerPool) workerFunc(worker *Worker) {

    for f := range worker.fn {
        if f == nil {
            break
        }

        f()

        if !wp.release(worker) {
            break
        }
    }

    wp.lock.Lock()
    wp.cur_worker_count--
    wp.lock.Unlock()

}

func (wp *WorkerPool) release(worker *Worker) bool {

	worker.last_used_time = time.Now().Unix()

    wp.lock.Lock()
    if wp.mustStop {
        wp.lock.Unlock()
        return false
    }
    wp.readyQue = append(wp.readyQue, worker)
    wp.lock.Unlock()

    return true
}

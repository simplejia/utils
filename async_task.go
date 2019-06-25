package utils

import (
	"sync/atomic"
	"time"
)

var (
	AsyncTaskChan = make(chan struct{}, 0)
	asyncTaskNum  = new(int32)
)

func AsyncTaskEnter() {
	atomic.AddInt32(asyncTaskNum, 1)
}

func AsyncTaskExit() {
	atomic.AddInt32(asyncTaskNum, -1)
}

func AsyncTaskShutdown(timeout time.Duration) {
	if timeout == 0 {
		return
	}

	n := atomic.LoadInt32(asyncTaskNum)
	if n == 0 {
		return
	}

	close(AsyncTaskChan)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timer.C:
			return
		case <-tick.C:
			if atomic.LoadInt32(asyncTaskNum) == 0 {
				return
			}
		}
	}
}

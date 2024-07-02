package waitgroup

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type WaitGroupTimeout struct {
	sync.WaitGroup
	start        chan struct{}
	waitForStart bool
}

func New(waitForStarts ...bool) *WaitGroupTimeout {
	var waitForStart bool
	if len(waitForStarts) > 0 {
		waitForStart = waitForStarts[0]
	}
	return &WaitGroupTimeout{
		start:        make(chan struct{}),
		waitForStart: waitForStart,
	}
}

func (wg *WaitGroupTimeout) Wrap(cb func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if wg.waitForStart {
			<-wg.start
		}
		cb()
	}()
}

func (wg *WaitGroupTimeout) Start() {
	select {
	case <-wg.start:
		return
	default:
		close(wg.start)
	}
}

func (wg *WaitGroupTimeout) WaitTimeout(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (wg *WaitGroupTimeout) Finished() bool {
	underfly := *(*waitGroup)(unsafe.Pointer(&wg.WaitGroup))
	var statep *uint64
	if uintptr(unsafe.Pointer(&underfly.state1))%8 == 0 {
		statep = (*uint64)(unsafe.Pointer(&underfly.state1))
	} else {
		statep = (*uint64)(unsafe.Pointer(&underfly.state1[1]))
	}
	state := atomic.LoadUint64(statep)
	v := int32(state >> 32)
	if v < 0 {
		panic("sync: negative WaitGroup counter")
	}
	if v > 0 {
		return false
	}

	return true
}

type noCopy struct{}
type waitGroup struct {
	noCopy noCopy

	// 64-bit value: high 32 bits are counter, low 32 bits are waiter count.
	// 64-bit atomic operations require 64-bit alignment, but 32-bit
	// compilers do not ensure it. So we allocate 12 bytes and then use
	// the aligned 8 bytes in them as state, and the other 4 as storage
	// for the sema.
	state1 [3]uint32
}

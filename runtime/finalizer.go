package runtime

import (
	"fmt"
	"os"
	goruntime "runtime"
	"sync"
	"weak"
)

// __del__: an object whose class defines it is given a Go finalizer. When the
// collector finds the object unreachable, the finalizer queues the call, and
// the thread that holds the interpreter lock runs it at the next back-edge
// (bit 4 of breaker), as signal handlers run. At exit, the objects still
// alive have their __del__ called, as CPython does at shutdown. The timing is
// the collector's, not reference counting's: __del__ does not run at the
// moment the last reference goes away.

var (
	delMu      sync.Mutex
	delQueue   []func()
	delPending []func() (func(), bool) // weak entries for the run at exit
)

// TrackDel registers obj's __del__.
func TrackDel[T any](obj *T, del func(*T)) {
	done := new(bool)
	wp := weak.Make(obj)
	run := func(o *T) func() {
		return func() {
			if *done {
				return
			}
			*done = true
			if e := Catch(func() { del(o) }); e != nil {
				fmt.Fprintf(os.Stderr, "Exception ignored while calling deallocator %s:\n%s", Repr(any(o)), report(e, map[*PyError]bool{}))
			}
		}
	}
	goruntime.SetFinalizer(obj, func(o *T) {
		delMu.Lock()
		delQueue = append(delQueue, run(o))
		delMu.Unlock()
		breaker.Or(4)
	})
	delMu.Lock()
	delPending = append(delPending, func() (func(), bool) {
		if o := wp.Value(); o != nil {
			return run(o), true
		}
		return nil, false
	})
	delMu.Unlock()
}

// runDelQueue calls the __del__ methods the collector queued.
func runDelQueue() {
	breaker.And(^int32(4))
	for {
		delMu.Lock()
		q := delQueue
		delQueue = nil
		delMu.Unlock()
		if len(q) == 0 {
			return
		}
		for _, f := range q {
			f()
		}
	}
}

// runDelAtExit calls __del__ for the objects still alive at the end.
func runDelAtExit() {
	runDelQueue()
	delMu.Lock()
	pending := delPending
	delPending = nil
	delMu.Unlock()
	for _, p := range pending {
		if f, alive := p(); alive {
			f()
		}
	}
}

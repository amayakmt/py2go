package runtime

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// The global interpreter lock. One goroutine runs Python code at a time. The
// threads that want it stand in a queue and get it in order, and a thread
// that wakes another reserves the woken thread's place when it wakes it, so
// threads run in the order they were made runnable, as they do in CPython
// when they are woken one after the other.
//
// The lock is given up when a thread blocks (sleep, locks, joins, reads) and,
// at the switch interval, when another thread is waiting: loops and calls of
// the generated code ask Yield whether it is time.

const mainThreadIdent int64 = 140737354129216

var (
	gilMu     sync.Mutex
	gilHeld   = true // the main thread has it from the start
	gilQ      []*gilTicket
	gilWaits  atomic.Int32
	breaker   atomic.Int32 // bit 1: a thread should yield; bit 2: a signal is pending
	curThread = mainThreadIdent
	threadSeq atomic.Int64
	liveCount atomic.Int64
	switchNs  atomic.Int64
	tickerOn  atomic.Bool

	shutdownHook any
)

func init() {
	switchNs.Store(int64(5 * time.Millisecond))
}

// gilTicket is a place in the queue for the lock.
type gilTicket struct{ ch chan struct{} }

// gilReserve takes a place in the queue, or the lock itself when it is free.
func gilReserve() *gilTicket {
	t := &gilTicket{ch: make(chan struct{}, 1)}
	gilMu.Lock()
	if !gilHeld {
		gilHeld = true
		t.ch <- struct{}{}
	} else {
		gilQ = append(gilQ, t)
		gilWaits.Store(int32(len(gilQ)))
	}
	gilMu.Unlock()
	return t
}

// wait returns when the ticket has the lock.
func (t *gilTicket) wait() { <-t.ch }

// gilRelease gives the lock to the first in the queue.
func gilRelease() {
	gilMu.Lock()
	if len(gilQ) > 0 {
		t := gilQ[0]
		gilQ = gilQ[1:]
		gilWaits.Store(int32(len(gilQ)))
		t.ch <- struct{}{}
	} else {
		gilHeld = false
	}
	gilMu.Unlock()
}

// Yield is called at loop back-edges. It costs one atomic load unless a
// thread wants the lock or a signal arrived.
func Yield() {
	if breaker.Load() != 0 {
		yieldSlow()
	}
}

func yieldSlow() {
	b := breaker.Load()
	if b&2 != 0 && curThread == mainThreadIdent {
		runSignalHandlers()
	}
	if b&4 != 0 {
		runDelQueue()
	}
	if b&1 != 0 {
		breaker.And(^int32(1))
		if gilWaits.Load() > 0 {
			me := curThread
			gilRelease()
			t := gilReserve()
			t.wait()
			curThread = me
		}
	}
}

// Blocking runs f with the lock released, for a call that waits.
func Blocking(f func()) {
	me := curThread
	gilRelease()
	defer func() {
		t := gilReserve()
		t.wait()
		curThread = me
	}()
	f()
}

func ensureTicker() {
	if !tickerOn.CompareAndSwap(false, true) {
		return
	}
	go func() {
		for {
			time.Sleep(time.Duration(switchNs.Load()))
			if gilWaits.Load() > 0 {
				breaker.Or(1)
			}
		}
	}()
}

// StartThread runs fn(*args, **kw) in a new thread and returns its ident.
func StartThread(fn any, args []any, kw []KW) int64 {
	id := mainThreadIdent + 1 + threadSeq.Add(1)*4096
	ensureTicker()
	liveCount.Add(1)
	ticket := gilReserve()
	go func() {
		ticket.wait()
		curThread = id
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(*PyError); ok {
					if !IsSubclass(e.Type, "SystemExit") {
						fmt.Fprintf(os.Stderr, "Exception ignored in thread started by: %s\n%s", Repr(fn), report(e, map[*PyError]bool{}))
					}
				} else {
					panic(r)
				}
			}
			liveCount.Add(-1)
			gilRelease()
		}()
		DynCallKw(fn, args, kw)
	}()
	return id
}

// sleepFor waits for d with the lock released, and runs the signal handlers
// of the main thread when a signal cuts the wait short.
func sleepFor(d time.Duration) {
	deadline := time.Now().Add(d)
	for {
		wake := signalWake()
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return
		}
		woken := false
		Blocking(func() {
			t := time.NewTimer(remaining)
			defer t.Stop()
			select {
			case <-t.C:
			case <-wake:
				woken = true
			}
		})
		if !woken {
			return
		}
		if curThread == mainThreadIdent {
			runSignalHandlers()
		}
	}
}

// ---- locks ----

// ThreadLock is _thread.lock. The threads that wait for it stand in a queue,
// and a release hands the lock to the first of them.
type ThreadLock struct {
	mu      sync.Mutex
	locked  bool
	waiters []*lockWaiter
}

type lockWaiter struct {
	ready  chan struct{}
	ticket *gilTicket // the place in the queue for the GIL, taken when the lock was handed over
	handed bool
}

func NewThreadLock() *ThreadLock { return &ThreadLock{} }

func (l *ThreadLock) PyClassName() string { return "lock" }
func (l *ThreadLock) PyRepr() string {
	state := "unlocked"
	if l.Locked() {
		state = "locked"
	}
	return "<" + state + " _thread.lock object at " + addr(l) + ">"
}

func (l *ThreadLock) Locked() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.locked
}

// acquire takes the lock; timeout < 0 waits for ever.
func (l *ThreadLock) acquire(blocking bool, timeout float64) bool {
	l.mu.Lock()
	if !l.locked {
		l.locked = true
		l.mu.Unlock()
		return true
	}
	l.mu.Unlock()
	if !blocking {
		return false
	}
	var deadline time.Time
	if timeout >= 0 {
		deadline = time.Now().Add(time.Duration(timeout * float64(time.Second)))
	}
	for {
		w := &lockWaiter{ready: make(chan struct{})}
		l.mu.Lock()
		if !l.locked {
			l.locked = true
			l.mu.Unlock()
			return true
		}
		l.waiters = append(l.waiters, w)
		l.mu.Unlock()

		me := curThread
		wake := signalWake()
		var timer <-chan time.Time
		if timeout >= 0 {
			t := time.NewTimer(time.Until(deadline))
			defer t.Stop()
			timer = t.C
		}
		gilRelease()
		timedOut, woken := false, false
		select {
		case <-w.ready:
		case <-timer:
			timedOut = true
		case <-wake:
			woken = true
		}
		if timedOut || woken {
			l.mu.Lock()
			if !w.handed {
				for i, x := range l.waiters {
					if x == w {
						l.waiters = append(l.waiters[:i], l.waiters[i+1:]...)
						break
					}
				}
				l.mu.Unlock()
				t := gilReserve()
				t.wait()
				curThread = me
				if timedOut {
					return false
				}
				if me == mainThreadIdent {
					runSignalHandlers()
				}
				continue
			}
			l.mu.Unlock()
		}
		w.ticket.wait()
		curThread = me
		return true
	}
}

func (l *ThreadLock) release() {
	l.mu.Lock()
	if !l.locked {
		l.mu.Unlock()
		panic(newError("RuntimeError", "release unlocked lock"))
	}
	if len(l.waiters) > 0 {
		w := l.waiters[0]
		l.waiters = l.waiters[1:]
		w.handed = true
		w.ticket = gilReserve()
		l.mu.Unlock()
		close(w.ready)
		return
	}
	l.locked = false
	l.mu.Unlock()
}

// lockArgs reads acquire(blocking=True, timeout=-1).
func lockArgs(args []any, kw []KW) (blocking bool, timeout float64) {
	blocking, timeout = true, -1
	if len(args) > 0 {
		blocking = DynBool(args[0])
	}
	if len(args) > 1 {
		timeout = mathFloat(args[1])
	}
	for _, k := range kw {
		switch k.Name {
		case "blocking":
			blocking = DynBool(k.Val)
		case "timeout":
			timeout = mathFloat(k.Val)
		}
	}
	if !blocking && timeout != -1 {
		panic(valueError("can't specify a timeout for a non-blocking call"))
	}
	if timeout < 0 && timeout != -1 {
		panic(valueError("timeout value must be a non-negative number"))
	}
	return blocking, timeout
}

var lockMethodNames = names("acquire release locked acquire_lock release_lock locked_lock __enter__ __exit__ _at_fork_reinit")

func (l *ThreadLock) PyDynTable() map[string]bool { return lockMethodNames }
func (l *ThreadLock) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "acquire", "acquire_lock", "__enter__":
		return l.acquire(lockArgs(args, kw)), true
	case "release", "release_lock":
		l.release()
		return nil, true
	case "__exit__":
		l.release()
		return nil, true
	case "locked", "locked_lock":
		return l.Locked(), true
	case "_at_fork_reinit":
		l.mu.Lock()
		l.locked, l.waiters = false, nil
		l.mu.Unlock()
		return nil, true
	}
	return nil, false
}

// ThreadRLock is _thread.RLock.
type ThreadRLock struct {
	lock  *ThreadLock
	owner int64
	count int64
}

func NewThreadRLock() *ThreadRLock { return &ThreadRLock{lock: NewThreadLock()} }

func (r *ThreadRLock) PyClassName() string { return "RLock" }
func (r *ThreadRLock) PyRepr() string {
	state := "unlocked"
	if r.count > 0 {
		state = "locked"
	}
	return fmt.Sprintf("<%s _thread.RLock object owner=%d count=%d at %s>", state, r.owner, r.count, addr(r))
}

func (r *ThreadRLock) acquire(blocking bool, timeout float64) bool {
	if r.count > 0 && r.owner == curThread {
		r.count++
		return true
	}
	if !r.lock.acquire(blocking, timeout) {
		return false
	}
	r.owner, r.count = curThread, 1
	return true
}

func (r *ThreadRLock) release() {
	if r.count == 0 || r.owner != curThread {
		panic(newError("RuntimeError", "cannot release un-acquired lock"))
	}
	r.count--
	if r.count == 0 {
		r.owner = 0
		r.lock.release()
	}
}

var rlockMethodNames = names("acquire release _is_owned _release_save _acquire_restore __enter__ __exit__ _recursion_count")

func (r *ThreadRLock) PyDynTable() map[string]bool { return rlockMethodNames }
func (r *ThreadRLock) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "acquire", "__enter__":
		return r.acquire(lockArgs(args, kw)), true
	case "release", "__exit__":
		r.release()
		return nil, true
	case "_is_owned":
		return r.count > 0 && r.owner == curThread, true
	case "_recursion_count":
		if r.owner == curThread {
			return r.count, true
		}
		return int64(0), true
	case "_release_save":
		if r.count == 0 {
			panic(newError("RuntimeError", "cannot release un-acquired lock"))
		}
		count, owner := r.count, r.owner
		r.count, r.owner = 0, 0
		r.lock.release()
		return NewVarTuple[any](count, owner), true
	case "_acquire_restore":
		countArgs("_acquire_restore", args, 1, 1)
		t := tupleInts(args[0])
		r.lock.acquire(true, -1)
		r.count, r.owner = t[0], t[1]
		return nil, true
	}
	return nil, false
}

func init() {
	RegisterNative("thread.start_new_thread", func(args []any, kw []KW) any {
		nativeArgs("start_new_thread", args, 2, 3)
		var fkw []KW
		if len(args) == 3 {
			if d, ok := Unbox(args[2]).(*Dict[any, any]); ok {
				for c := d.Iter(); c.Next(); {
					v := d.Get(c.Value())
					fkw = append(fkw, KW{Name: Str(c.Value()), Val: v})
				}
			}
		}
		var fargs []any
		for c := DynIter(args[1]).Iter(); c.Next(); {
			fargs = append(fargs, c.Value())
		}
		return StartThread(args[0], fargs, fkw)
	})
	RegisterNative("thread.get_ident", func(args []any, kw []KW) any { return curThread })
	RegisterNative("thread.main_ident", func(args []any, kw []KW) any { return mainThreadIdent })
	RegisterNative("thread.count", func(args []any, kw []KW) any { return liveCount.Load() })
	RegisterNative("thread.allocate_lock", func(args []any, kw []KW) any { return NewThreadLock() })
	RegisterNative("thread.RLock", func(args []any, kw []KW) any { return NewThreadRLock() })
	RegisterNative("thread.setswitchinterval", func(args []any, kw []KW) any {
		nativeArgs("setswitchinterval", args, 1, 1)
		f := mathFloat(args[0])
		if f <= 0 {
			panic(valueError("switch interval must be strictly positive"))
		}
		switchNs.Store(int64(f * 1e9))
		return nil
	})
	RegisterNative("thread.getswitchinterval", func(args []any, kw []KW) any {
		return float64(switchNs.Load()) / 1e9
	})
	RegisterNative("thread.set_shutdown_hook", func(args []any, kw []KW) any {
		nativeArgs("set_shutdown_hook", args, 1, 1)
		shutdownHook = args[0]
		return nil
	})
	RegisterNative("thread.sleep0", func(args []any, kw []KW) any {
		Blocking(func() { time.Sleep(0) })
		return nil
	})
	RegisterNative("atexit.register", func(args []any, kw []KW) any {
		if len(args) < 1 {
			panic(newError("TypeError", "register() takes at least 1 argument (0 given)"))
		}
		if !callableValue(args[0]) {
			panic(newError("TypeError", "the first argument must be callable"))
		}
		exitFuncs = append(exitFuncs, exitFunc{fn: args[0], args: args[1:], kw: kw})
		return args[0]
	})
	RegisterNative("atexit.unregister", func(args []any, kw []KW) any {
		nativeArgs("unregister", args, 1, 1)
		kept := exitFuncs[:0]
		for _, f := range exitFuncs {
			if !Eq(f.fn, args[0]) {
				kept = append(kept, f)
			}
		}
		exitFuncs = kept
		return nil
	})
	RegisterNative("atexit._clear", func(args []any, kw []KW) any { exitFuncs = nil; return nil })
	RegisterNative("atexit._ncallbacks", func(args []any, kw []KW) any { return int64(len(exitFuncs)) })
	RegisterNative("atexit._run_exitfuncs", func(args []any, kw []KW) any { runAtExit(); return nil })
}

type exitFunc struct {
	fn   any
	args []any
	kw   []KW
}

var exitFuncs []exitFunc

func callableValue(v any) bool { return DynCallable(v) }

// runAtExit calls the registered functions, last first, as atexit does.
func runAtExit() {
	for len(exitFuncs) > 0 {
		f := exitFuncs[len(exitFuncs)-1]
		exitFuncs = exitFuncs[:len(exitFuncs)-1]
		if e := Catch(func() { DynCallKw(f.fn, f.args, f.kw) }); e != nil {
			if IsSubclass(e.Type, "SystemExit") {
				continue
			}
			fmt.Fprintf(os.Stderr, "Exception ignored in atexit callback %s:\n%s", Repr(f.fn), report(e, map[*PyError]bool{}))
		}
	}
}

// runExitHooks is what happens when the program ends: the threads that are not
// daemons are waited for, then the functions of atexit run.
func runExitHooks() {
	if shutdownHook != nil {
		h := shutdownHook
		shutdownHook = nil
		if e := Catch(func() { DynCall(h) }); e != nil && !IsSubclass(e.Type, "SystemExit") {
			fmt.Fprintf(os.Stderr, "Exception ignored in: %s\n%s", Repr(h), report(e, map[*PyError]bool{}))
		}
	}
	runAtExit()
	runDelAtExit()
	FlushAll()
}

// ThreadLocal is threading.local: its attributes are per thread.
type ThreadLocal struct{ data map[int64]map[string]any }

func (t *ThreadLocal) PyClassName() string { return "local" }
func (t *ThreadLocal) PyRepr() string      { return "<_thread._local object at " + addr(t) + ">" }

func (t *ThreadLocal) mine(create bool) map[string]any {
	m := t.data[curThread]
	if m == nil && create {
		if t.data == nil {
			t.data = map[int64]map[string]any{}
		}
		m = map[string]any{}
		t.data[curThread] = m
	}
	return m
}

func (t *ThreadLocal) PyGetAttr(name string) (any, bool) {
	if name == "__dict__" {
		d := NewDict[any, any]()
		for k, v := range t.mine(false) {
			d.Set(k, v)
		}
		return d, true
	}
	v, ok := t.mine(false)[name]
	return v, ok
}

func (t *ThreadLocal) PySetAttr(name string, v any) bool {
	t.mine(true)[name] = v
	return true
}

func (t *ThreadLocal) PyDelAttr(name string) bool {
	m := t.mine(false)
	if _, ok := m[name]; !ok {
		return false
	}
	delete(m, name)
	return true
}

func init() {
	RegisterNative("thread.local", func(args []any, kw []KW) any { return &ThreadLocal{} })
}

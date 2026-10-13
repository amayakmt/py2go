package runtime

import (
	"sync"
	"syscall"
	"testing"
	"time"
)

// The tests that start threads run them with the lock like the generated
// code does: the test goroutine is the main thread and holds the lock.

func TestThreadLock(t *testing.T) {
	l := NewThreadLock()
	if l.Locked() || !l.acquire(true, -1) || !l.Locked() {
		t.Fatal("acquire of a free lock")
	}
	if l.acquire(false, -1) {
		t.Error("a non-blocking acquire of a held lock succeeds")
	}
	if l.acquire(true, 0.01) {
		t.Error("an acquire with a timeout of a held lock succeeds")
	}
	l.release()
	if l.Locked() {
		t.Error("released lock is locked")
	}
	if got := catch(func() { l.release() }); got != "RuntimeError: release unlocked lock" {
		t.Errorf("release of a free lock: %s", got)
	}
	if got := catch(func() { lockArgs([]any{false, 1.0}, nil) }); got != "ValueError: can't specify a timeout for a non-blocking call" {
		t.Errorf("got %s", got)
	}
	if got := catch(func() { lockArgs(nil, []KW{{Name: "timeout", Val: -5.0}}) }); got != "ValueError: timeout value must be a non-negative number" {
		t.Errorf("got %s", got)
	}
}

func TestThreadLockHandsOverInOrder(t *testing.T) {
	l := NewThreadLock()
	l.acquire(true, -1)
	var mu sync.Mutex
	var order []int
	start := func(i int) {
		StartThread(func() any {
			l.acquire(true, -1)
			mu.Lock()
			order = append(order, i)
			mu.Unlock()
			l.release()
			return NoneV
		}, nil, nil)
	}
	for i := 0; i < 4; i++ {
		start(i)
		// let the thread reach the lock before the next one starts
		for {
			l.mu.Lock()
			n := len(l.waiters)
			l.mu.Unlock()
			if n == i+1 {
				break
			}
			Blocking(func() { time.Sleep(time.Millisecond) })
		}
	}
	l.release()
	Blocking(func() { time.Sleep(50 * time.Millisecond) })
	mu.Lock()
	defer mu.Unlock()
	for i, got := range order {
		if got != i {
			t.Fatalf("threads took the lock in the order %v", order)
		}
	}
	if len(order) != 4 {
		t.Fatalf("only %v ran", order)
	}
}

func TestThreadRLock(t *testing.T) {
	r := NewThreadRLock()
	if !r.acquire(true, -1) || !r.acquire(true, -1) {
		t.Fatal("reentrant acquire")
	}
	r.release()
	if r.count != 1 {
		t.Errorf("count %d", r.count)
	}
	r.release()
	if got := catch(func() { r.release() }); got != "RuntimeError: cannot release un-acquired lock" {
		t.Errorf("got %s", got)
	}
	r.acquire(true, -1)
	r.acquire(true, -1)
	state, _ := r.DynMethod("_release_save", nil, nil)
	if r.count != 0 || r.lock.Locked() {
		t.Error("_release_save leaves the lock held")
	}
	r.DynMethod("_acquire_restore", []any{state}, nil)
	if r.count != 2 || r.owner != curThread {
		t.Errorf("_acquire_restore: count %d owner %d", r.count, r.owner)
	}
}

func TestThreadLocal(t *testing.T) {
	l := &ThreadLocal{}
	l.PySetAttr("x", int64(1))
	done := make(chan any)
	StartThread(func() any {
		_, seen := l.PyGetAttr("x")
		l.PySetAttr("x", int64(2))
		v, _ := l.PyGetAttr("x")
		done <- []any{seen, v}
		return NoneV
	}, nil, nil)
	var res any
	Blocking(func() { res = <-done })
	got := res.([]any)
	if got[0] != false || got[1] != int64(2) {
		t.Errorf("the other thread saw %v", got)
	}
	if v, _ := l.PyGetAttr("x"); v != int64(1) {
		t.Errorf("main thread has %v", v)
	}
	if !l.PyDelAttr("x") || l.PyDelAttr("x") {
		t.Error("delete")
	}
}

func TestSignalHandlers(t *testing.T) {
	var got []int64
	h := func(sig, frame any) any { got = append(got, sig.(int64)); return NoneV }
	sigMu.Lock()
	sigHandlers[int(syscall.SIGUSR1)] = any(h)
	sigMu.Unlock()
	deliverSignal(int(syscall.SIGUSR1))
	if breaker.Load()&2 == 0 {
		t.Error("a signal sets the flag")
	}
	Yield()
	if len(got) != 1 || got[0] != int64(syscall.SIGUSR1) {
		t.Errorf("handler calls %v", got)
	}
	if breaker.Load()&2 != 0 {
		t.Error("the flag is cleared")
	}
	deliverSignal(int(syscall.SIGINT))
	if got := catch(func() { runSignalHandlers() }); got != "KeyboardInterrupt" {
		t.Errorf("SIGINT: %q", got)
	}
}

func TestAtExitRunsLastFirst(t *testing.T) {
	var order []string
	saved := exitFuncs
	exitFuncs = nil
	defer func() { exitFuncs = saved }()
	add := func(name string) {
		exitFuncs = append(exitFuncs, exitFunc{fn: func() any { order = append(order, name); return NoneV }})
	}
	add("a")
	add("b")
	add("c")
	runAtExit()
	if len(order) != 3 || order[0] != "c" || order[2] != "a" {
		t.Errorf("order %v", order)
	}
}

func TestSameFunc(t *testing.T) {
	f := func() any { return nil }
	g := func() any { return nil }
	if r, isFunc := sameFunc(f, f); !isFunc || !r {
		t.Error("a function is the same as itself")
	}
	if r, _ := sameFunc(f, g); r {
		t.Error("two closures are different")
	}
	if _, isFunc := sameFunc(1, 2); isFunc {
		t.Error("ints are not functions")
	}
	if !Eq(&Callable{Fn: f}, &Callable{Fn: f}) {
		t.Error("callables of one function are equal")
	}
}

func TestCoroutineBasics(t *testing.T) {
	g := NewGen(func(y *Yielder[any, any]) any {
		v := y.Yield("waiting")
		return v
	}).AsCoro("co")
	if TypeName(g) != "coroutine" {
		t.Errorf("type %s", TypeName(g))
	}
	if got := catch(func() { DynCallMethod(g, "send", []any{int64(1)}, nil) }); got != "TypeError: can't send non-None value to a just-started coroutine" {
		t.Errorf("send of a value: %s", got)
	}
	if v := DynCallMethod(g, "send", []any{nil}, nil); v != "waiting" {
		t.Errorf("first send gave %v", v)
	}
	got := catch(func() { DynCallMethod(g, "send", []any{int64(7)}, nil) })
	if got != "StopIteration: 7" && got != "StopIteration" {
		t.Errorf("finish: %s", got)
	}
	if got := catch(func() { DynCallMethod(g, "send", []any{nil}, nil) }); got != "RuntimeError: cannot reuse already awaited coroutine" {
		t.Errorf("reuse: %s", got)
	}
}

func TestAwaitDelegates(t *testing.T) {
	inner := NewGen(func(y *Yielder[any, any]) any {
		y.Yield("a")
		return int64(5)
	}).AsCoro("inner")
	outer := NewGen(func(y *Yielder[any, any]) any {
		r := AwaitObj(y, inner)
		return r
	}).AsCoro("outer")
	if v := outer.Send(nil); v != "a" {
		t.Errorf("yielded %v", v)
	}
	got := catch(func() { outer.Send(nil) })
	if got != "StopIteration: 5" {
		t.Errorf("finished with %s", got)
	}
	bad := NewGen(func(y *Yielder[any, any]) any { return AwaitObj(y, int64(3)) }).AsCoro("bad")
	if got := catch(func() { bad.Send(nil) }); got != "TypeError: object int can't be used in 'await' expression" {
		t.Errorf("await of an int: %s", got)
	}
}

func TestAsyncGenerator(t *testing.T) {
	ag := NewAsyncGen(NewGen(func(y *Yielder[any, any]) any {
		AGYield(y, int64(1))
		y.Yield("future") // an await inside the generator passes through
		AGYield(y, int64(2))
		return NoneV
	}))
	drive := func(c *Gen[any, any, any]) (any, string) {
		var out any
		msg := catch(func() {
			for {
				v := c.Send(nil)
				if v == "future" {
					continue
				}
				out = v
			}
		})
		return out, msg
	}
	next := func() (any, string) {
		c, _ := ag.DynMethod("__anext__", nil, nil)
		return drive(c.(*Gen[any, any, any]))
	}
	for _, want := range []int64{1, 2} {
		_, msg := next()
		if msg != "StopIteration: "+itoa(int(want)) {
			t.Errorf("item %d: %s", want, msg)
		}
	}
	if _, msg := next(); msg != "StopAsyncIteration" {
		t.Errorf("end: %s", msg)
	}
	if ag.PyClassName() != "async_generator" {
		t.Error("type name")
	}
}

func TestYieldFastPath(t *testing.T) {
	breaker.Store(0)
	Yield() // must not block or panic
}

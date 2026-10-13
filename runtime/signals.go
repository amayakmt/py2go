package runtime

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Signal handlers are Python callables that run in the main thread, between
// calls: the arrival of a signal sets a flag the loops and the blocking calls
// look at (Yield, sleepFor, lock waits).

const (
	sigDfl = 0
	sigIgn = 1
)

var (
	sigMu       sync.Mutex
	sigHandlers = map[int]any{}
	sigPending  = map[int]bool{}
	sigWakeCh   = make(chan struct{})
	sigNotify   chan os.Signal
	sigNotified = map[int]bool{}
	alarmTimer  *time.Timer
	alarmDue    time.Time
	itimerStop  chan struct{}
)

func init() {
	// Ctrl-C raises KeyboardInterrupt from the start.
	sigHandlers[int(syscall.SIGINT)] = "default_int_handler"
	notifySig(int(syscall.SIGINT))
}

func notifySig(sig int) {
	sigMu.Lock()
	defer sigMu.Unlock()
	if sigNotify == nil {
		sigNotify = make(chan os.Signal, 64)
		go func() {
			for s := range sigNotify {
				if n, ok := s.(syscall.Signal); ok {
					deliverSignal(int(n))
				}
			}
		}()
	}
	if !sigNotified[sig] {
		sigNotified[sig] = true
		signal.Notify(sigNotify, syscall.Signal(sig))
	}
}

// deliverSignal records a signal and wakes the main thread.
func deliverSignal(sig int) {
	sigMu.Lock()
	sigPending[sig] = true
	old := sigWakeCh
	sigWakeCh = make(chan struct{})
	sigMu.Unlock()
	breaker.Or(2)
	close(old)
}

func signalWake() chan struct{} {
	sigMu.Lock()
	defer sigMu.Unlock()
	return sigWakeCh
}

// runSignalHandlers calls the Python handlers of the signals that arrived.
func runSignalHandlers() {
	breaker.And(^int32(2))
	for {
		sigMu.Lock()
		sig := 0
		for s := range sigPending {
			if sig == 0 || s < sig {
				sig = s
			}
		}
		if sig == 0 {
			sigMu.Unlock()
			return
		}
		delete(sigPending, sig)
		h := sigHandlers[sig]
		sigMu.Unlock()
		switch x := h.(type) {
		case nil:
		case string:
			if x == "default_int_handler" {
				panic(newError("KeyboardInterrupt", ""))
			}
		case int64:
			if x == sigDfl && sig != int(syscall.SIGINT) {
				signal.Reset(syscall.Signal(sig))
				syscall.Kill(os.Getpid(), syscall.Signal(sig))
			}
		default:
			DynCall(h, int64(sig), NoneV)
		}
	}
}

func checkSig(sig int64) int {
	if sig < 1 || sig > 64 {
		panic(valueError("signal number out of range"))
	}
	return int(sig)
}

func init() {
	RegisterNative("signal.signal", func(args []any, kw []KW) any {
		nativeArgs("signal", args, 2, 2)
		sig := checkSig(DynInt(args[0]))
		if curThread != mainThreadIdent {
			panic(valueError("signal only works in main thread of the main interpreter"))
		}
		if sig == int(syscall.SIGKILL) || sig == int(syscall.SIGSTOP) {
			e := newError("OSError", "[Errno 22] Invalid argument")
			e.Attrs = map[string]any{"errno": int64(22), "strerror": "Invalid argument"}
			panic(e)
		}
		h := Unbox(args[1])
		sigMu.Lock()
		prev, ok := sigHandlers[sig]
		sigMu.Unlock()
		if !ok {
			prev = int64(sigDfl)
		}
		switch x := h.(type) {
		case int64:
			if x != sigDfl && x != sigIgn {
				panic(newError("TypeError", "signal handler must be signal.SIG_IGN, signal.SIG_DFL, or a callable object"))
			}
		default:
			if !DynCallable(h) {
				panic(newError("TypeError", "signal handler must be signal.SIG_IGN, signal.SIG_DFL, or a callable object"))
			}
		}
		sigMu.Lock()
		sigHandlers[sig] = h
		sigMu.Unlock()
		switch x := h.(type) {
		case int64:
			if x == sigIgn {
				signal.Ignore(syscall.Signal(sig))
				sigMu.Lock()
				delete(sigNotified, sig)
				sigMu.Unlock()
			} else {
				signal.Reset(syscall.Signal(sig))
				sigMu.Lock()
				delete(sigNotified, sig)
				sigMu.Unlock()
				if sig == int(syscall.SIGINT) {
					sigMu.Lock()
					sigHandlers[sig] = "default_int_handler"
					sigMu.Unlock()
					notifySig(sig)
				}
			}
		default:
			notifySig(sig)
		}
		if s, ok := prev.(string); ok && s == "default_int_handler" {
			return "default_int_handler"
		}
		return prev
	})
	RegisterNative("signal.getsignal", func(args []any, kw []KW) any {
		nativeArgs("getsignal", args, 1, 1)
		sig := checkSig(DynInt(args[0]))
		sigMu.Lock()
		defer sigMu.Unlock()
		if h, ok := sigHandlers[sig]; ok {
			return h
		}
		return int64(sigDfl)
	})
	RegisterNative("signal.raise_signal", func(args []any, kw []KW) any {
		nativeArgs("raise_signal", args, 1, 1)
		sig := checkSig(DynInt(args[0]))
		sigMu.Lock()
		h, ok := sigHandlers[sig]
		sigMu.Unlock()
		if !ok || h == any(int64(sigDfl)) {
			signal.Reset(syscall.Signal(sig))
			syscall.Kill(os.Getpid(), syscall.Signal(sig))
			time.Sleep(50 * time.Millisecond)
			return nil
		}
		if h == any(int64(sigIgn)) {
			return nil
		}
		deliverSignal(sig)
		if curThread == mainThreadIdent {
			runSignalHandlers()
		}
		return nil
	})
	RegisterNative("signal.kill", func(args []any, kw []KW) any {
		nativeArgs("kill", args, 2, 2)
		pid, sig := int(DynInt(args[0])), int(DynInt(args[1]))
		if pid == os.Getpid() {
			sigMu.Lock()
			h, ok := sigHandlers[sig]
			sigMu.Unlock()
			if ok && h != any(int64(sigDfl)) && h != any(int64(sigIgn)) {
				deliverSignal(sig)
				if curThread == mainThreadIdent {
					runSignalHandlers()
				}
				return nil
			}
		}
		if err := syscall.Kill(pid, syscall.Signal(sig)); err != nil {
			panic(osError(err, ""))
		}
		return nil
	})
	RegisterNative("signal.alarm", func(args []any, kw []KW) any {
		nativeArgs("alarm", args, 1, 1)
		secs := DynInt(args[0])
		prev := int64(0)
		sigMu.Lock()
		if alarmTimer != nil {
			if alarmTimer.Stop() {
				prev = int64((time.Until(alarmDue) + time.Second - 1) / time.Second)
			}
			alarmTimer = nil
		}
		if secs > 0 {
			alarmDue = time.Now().Add(time.Duration(secs) * time.Second)
			alarmTimer = time.AfterFunc(time.Duration(secs)*time.Second, func() { fireAlarm(int(syscall.SIGALRM)) })
		}
		sigMu.Unlock()
		return prev
	})
	RegisterNative("signal.setitimer", func(args []any, kw []KW) any {
		nativeArgs("setitimer", args, 2, 3)
		secs := mathFloat(args[1])
		interval := 0.0
		if len(args) > 2 {
			interval = mathFloat(args[2])
		}
		sigMu.Lock()
		if alarmTimer != nil {
			alarmTimer.Stop()
			alarmTimer = nil
		}
		if itimerStop != nil {
			close(itimerStop)
			itimerStop = nil
		}
		sigMu.Unlock()
		if secs > 0 {
			d := time.Duration(secs * float64(time.Second))
			sigMu.Lock()
			alarmDue = time.Now().Add(d)
			stop := make(chan struct{})
			if interval > 0 {
				itimerStop = stop
			}
			alarmTimer = time.AfterFunc(d, func() {
				fireAlarm(int(syscall.SIGALRM))
				if interval > 0 {
					t := time.NewTicker(time.Duration(interval * float64(time.Second)))
					defer t.Stop()
					for {
						select {
						case <-t.C:
							fireAlarm(int(syscall.SIGALRM))
						case <-stop:
							return
						}
					}
				}
			})
			sigMu.Unlock()
		}
		return NewVarTuple[any](0.0, 0.0)
	})
	RegisterNative("os.getppid", func(args []any, kw []KW) any { return int64(os.Getppid()) })
	RegisterNative("os._exit", func(args []any, kw []KW) any {
		nativeArgs("_exit", args, 1, 1)
		os.Exit(int(DynInt(args[0])))
		return nil
	})
	RegisterNative("signal.getpid", func(args []any, kw []KW) any { return int64(os.Getpid()) })
	RegisterNative("signal.pause", func(args []any, kw []KW) any {
		for {
			wake := signalWake()
			Blocking(func() { <-wake })
			if curThread == mainThreadIdent {
				runSignalHandlers()
			}
			return nil
		}
	})
	RegisterNative("signal.default_int_handler", func(args []any, kw []KW) any {
		panic(newError("KeyboardInterrupt", ""))
	})
	RegisterNative("thread.interrupt_main", func(args []any, kw []KW) any {
		sig := int(syscall.SIGINT)
		if len(args) > 0 {
			sig = checkSig(DynInt(args[0]))
		}
		sigMu.Lock()
		h := sigHandlers[sig]
		sigMu.Unlock()
		if h == any(int64(sigDfl)) || h == any(int64(sigIgn)) || h == nil {
			return nil
		}
		deliverSignal(sig)
		return nil
	})
}

// fireAlarm delivers SIGALRM to a handler, or lets the default action kill
// the process.
func fireAlarm(sig int) {
	sigMu.Lock()
	h, ok := sigHandlers[sig]
	sigMu.Unlock()
	if ok && h != any(int64(sigDfl)) {
		if h != any(int64(sigIgn)) {
			deliverSignal(sig)
		}
		return
	}
	signal.Reset(syscall.Signal(sig))
	syscall.Kill(os.Getpid(), syscall.Signal(sig))
}

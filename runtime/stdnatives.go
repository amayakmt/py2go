package runtime

import (
	cryptoRand "crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
	"unsafe"
)

// The native parts of sys, os, time and copy.

var startTime = time.Now()

func strList(xs []string) *List[any] {
	l := NewList[any]()
	for _, x := range xs {
		l.Append(x)
	}
	return l
}

func init() {
	// ---- sys ----
	RegisterNative("sys.argv", func(args []any, kw []KW) any {
		argv := []string{os.Args[0]}
		if len(os.Args) > 1 {
			argv = append(argv, os.Args[1:]...)
		}
		return strList(argv)
	})
	RegisterNative("sys.stdout", func(args []any, kw []KW) any { return Stdout })
	RegisterNative("sys.stderr", func(args []any, kw []KW) any { return Stderr })
	RegisterNative("sys.stdin", func(args []any, kw []KW) any { return Stdin })
	RegisterNative("sys.getrefcount", func(args []any, kw []KW) any { return int64(2) })

	// ---- os ----
	RegisterNative("os.getcwd", func(args []any, kw []KW) any {
		wd, _ := os.Getwd()
		return wd
	})
	RegisterNative("os.environ", func(args []any, kw []KW) any {
		d := NewDict[any, any]()
		env := os.Environ()
		sort.Strings(env)
		for _, kv := range env {
			if k, v, ok := strings.Cut(kv, "="); ok {
				d.Set(k, v)
			}
		}
		return d
	})
	RegisterNative("os.getenv", func(args []any, kw []KW) any {
		nativeArgs("getenv", args, 1, 2)
		if v, ok := os.LookupEnv(Str(args[0])); ok {
			return v
		}
		if len(args) > 1 {
			return args[1]
		}
		return nil
	})
	RegisterNative("os.putenv", func(args []any, kw []KW) any {
		os.Setenv(Str(args[0]), Str(args[1]))
		return nil
	})
	RegisterNative("os.unsetenv", func(args []any, kw []KW) any {
		os.Unsetenv(Str(args[0]))
		return nil
	})
	RegisterNative("os.getpid", func(args []any, kw []KW) any { return int64(os.Getpid()) })
	RegisterNative("os.cpu_count", func(args []any, kw []KW) any { return int64(runtime.NumCPU()) })
	RegisterNative("os.listdir", func(args []any, kw []KW) any {
		dir := "."
		if len(args) > 0 && !IsNone(args[0]) {
			dir = Str(args[0])
		}
		ents, err := os.ReadDir(dir)
		if err != nil {
			panic(osError(err, dir))
		}
		var names []string
		for _, e := range ents {
			names = append(names, e.Name())
		}
		return strList(names)
	})
	RegisterNative("os.mkdir", func(args []any, kw []KW) any {
		path := Str(args[0])
		if err := os.Mkdir(path, 0o777); err != nil {
			panic(osError(err, path))
		}
		return nil
	})
	RegisterNative("os.makedirs", func(args []any, kw []KW) any {
		path := Str(args[0])
		exist := false
		if v, ok := kwValue(kw, "exist_ok"); ok {
			exist = Truth(v)
		}
		if len(args) > 2 {
			exist = Truth(args[2])
		}
		if st, err := os.Stat(path); err == nil && st.IsDir() {
			if exist {
				return nil
			}
			panic(osErrorCode("FileExistsError", 17, "File exists", path))
		}
		if err := os.MkdirAll(path, 0o777); err != nil {
			panic(osError(err, path))
		}
		return nil
	})
	RegisterNative("os.remove", func(args []any, kw []KW) any {
		path := Str(args[0])
		if err := os.Remove(path); err != nil {
			panic(osError(err, path))
		}
		return nil
	})
	RegisterNative("os.rmdir", func(args []any, kw []KW) any {
		path := Str(args[0])
		if err := os.Remove(path); err != nil {
			panic(osError(err, path))
		}
		return nil
	})
	RegisterNative("os.rename", func(args []any, kw []KW) any {
		a, b := Str(args[0]), Str(args[1])
		if err := os.Rename(a, b); err != nil {
			panic(osError(err, a))
		}
		return nil
	})
	RegisterNative("os.stat", func(args []any, kw []KW) any {
		path := Str(args[0])
		st, err := os.Stat(path)
		if err != nil {
			panic(osError(err, path))
		}
		mode := int64(st.Mode().Perm())
		if st.IsDir() {
			mode |= 0o040000
		} else {
			mode |= 0o100000
		}
		return NewVarTuple[any](mode, st.Size(), float64(st.ModTime().UnixNano())/1e9)
	})
	RegisterNative("os.path.exists", func(args []any, kw []KW) any {
		_, err := os.Stat(Str(args[0]))
		return err == nil
	})
	RegisterNative("os.path.isfile", func(args []any, kw []KW) any {
		st, err := os.Stat(Str(args[0]))
		return err == nil && st.Mode().IsRegular()
	})
	RegisterNative("os.path.isdir", func(args []any, kw []KW) any {
		st, err := os.Stat(Str(args[0]))
		return err == nil && st.IsDir()
	})
	RegisterNative("os.path.islink", func(args []any, kw []KW) any {
		st, err := os.Lstat(Str(args[0]))
		return err == nil && st.Mode()&os.ModeSymlink != 0
	})
	RegisterNative("os.path.realpath", func(args []any, kw []KW) any {
		p := Str(args[0])
		abs, err := filepath.Abs(p)
		if err != nil {
			return p
		}
		if r, err := filepath.EvalSymlinks(abs); err == nil {
			return r
		}
		return abs
	})
	RegisterNative("os.path.expanduser", func(args []any, kw []KW) any {
		p := Str(args[0])
		if p == "~" || strings.HasPrefix(p, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				return home + p[1:]
			}
		}
		return p
	})

	// ---- time ----
	RegisterNative("time.time", func(args []any, kw []KW) any { return float64(time.Now().UnixNano()) / 1e9 })
	RegisterNative("time.time_ns", func(args []any, kw []KW) any { return time.Now().UnixNano() })
	RegisterNative("time.perf_counter", func(args []any, kw []KW) any { return time.Since(startTime).Seconds() })
	RegisterNative("time.perf_counter_ns", func(args []any, kw []KW) any { return int64(time.Since(startTime)) })
	RegisterNative("time.process_time", func(args []any, kw []KW) any { return time.Since(startTime).Seconds() })
	RegisterNative("time.sleep", func(args []any, kw []KW) any {
		nativeArgs("sleep", args, 1, 1)
		secs := mathFloat(args[0])
		if secs < 0 {
			panic(valueError("sleep length must be non-negative"))
		}
		sleepFor(time.Duration(secs * float64(time.Second)))
		return nil
	})
	RegisterNative("time.gmtime", func(args []any, kw []KW) any { return timeTuple(args, time.UTC) })
	RegisterNative("time.localtime", func(args []any, kw []KW) any { return timeTuple(args, time.Local) })
	RegisterNative("time.mktime", func(args []any, kw []KW) any {
		t := tupleInts(args[0])
		return float64(time.Date(int(t[0]), time.Month(t[1]), int(t[2]), int(t[3]), int(t[4]), int(t[5]), 0, time.Local).Unix())
	})
	RegisterNative("time.timegm", func(args []any, kw []KW) any {
		t := tupleInts(args[0])
		return time.Date(int(t[0]), time.Month(t[1]), int(t[2]), int(t[3]), int(t[4]), int(t[5]), 0, time.UTC).Unix()
	})
	RegisterNative("time.strftime", func(args []any, kw []KW) any {
		t := tupleInts(args[1])
		loc := time.UTC
		if t[8] >= 0 {
			loc = time.Local
		}
		return strftime(Str(args[0]), time.Date(int(t[0]), time.Month(t[1]), int(t[2]), int(t[3]), int(t[4]), int(t[5]), 0, loc))
	})

	// ---- copy ----
	RegisterNative("copy.copy", func(args []any, kw []KW) any { return shallowCopy(args[0]) })
	RegisterNative("copy.deepcopy", func(args []any, kw []KW) any {
		return deepCopy(args[0], map[uintptr]reflect.Value{})
	})
}

func tupleInts(v any) []int64 {
	out := make([]int64, 0, 9)
	for c := DynIter(v).Iter(); c.Next(); {
		out = append(out, DynInt(c.Value()))
	}
	for len(out) < 9 {
		out = append(out, 0)
	}
	return out
}

// timeTuple is the nine fields of a struct_time for a time stamp.
func timeTuple(args []any, loc *time.Location) any {
	var t time.Time
	if len(args) == 0 || IsNone(args[0]) {
		t = time.Now()
	} else {
		secs := mathFloat(args[0])
		t = time.Unix(int64(secs), 0)
	}
	t = t.In(loc)
	wday := (int64(t.Weekday()) + 6) % 7
	dst := int64(0)
	if loc != time.UTC && t.IsDST() {
		dst = 1
	}
	if loc == time.UTC {
		dst = 0
	}
	return NewVarTuple[any](int64(t.Year()), int64(t.Month()), int64(t.Day()), int64(t.Hour()), int64(t.Minute()), int64(t.Second()), wday, int64(t.YearDay()), dst)
}

// strftime is the C library's strftime for the common conversions.
func strftime(format string, t time.Time) string {
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		i++
		switch format[i] {
		case 'Y':
			fmt.Fprintf(&b, "%d", t.Year())
		case 'y':
			fmt.Fprintf(&b, "%02d", t.Year()%100)
		case 'm':
			fmt.Fprintf(&b, "%02d", int(t.Month()))
		case 'd':
			fmt.Fprintf(&b, "%02d", t.Day())
		case 'e':
			fmt.Fprintf(&b, "%2d", t.Day())
		case 'H':
			fmt.Fprintf(&b, "%02d", t.Hour())
		case 'I':
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			fmt.Fprintf(&b, "%02d", h)
		case 'M':
			fmt.Fprintf(&b, "%02d", t.Minute())
		case 'S':
			fmt.Fprintf(&b, "%02d", t.Second())
		case 'p':
			if t.Hour() < 12 {
				b.WriteString("AM")
			} else {
				b.WriteString("PM")
			}
		case 'a':
			b.WriteString(t.Weekday().String()[:3])
		case 'A':
			b.WriteString(t.Weekday().String())
		case 'b', 'h':
			b.WriteString(t.Month().String()[:3])
		case 'B':
			b.WriteString(t.Month().String())
		case 'j':
			fmt.Fprintf(&b, "%03d", t.YearDay())
		case 'w':
			fmt.Fprintf(&b, "%d", int(t.Weekday()))
		case 'u':
			wd := int(t.Weekday())
			if wd == 0 {
				wd = 7
			}
			fmt.Fprintf(&b, "%d", wd)
		case 'Z':
			name, _ := t.Zone()
			b.WriteString(name)
		case 'z':
			b.WriteString(t.Format("-0700"))
		case 'c':
			b.WriteString(t.Format("Mon Jan _2 15:04:05 2006"))
		case 'x':
			b.WriteString(t.Format("01/02/06"))
		case 'X':
			b.WriteString(t.Format("15:04:05"))
		case 'D':
			b.WriteString(t.Format("01/02/06"))
		case 'F':
			b.WriteString(t.Format("2006-01-02"))
		case 'T':
			b.WriteString(t.Format("15:04:05"))
		case 'R':
			b.WriteString(t.Format("15:04"))
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(format[i])
		}
	}
	return b.String()
}

// ---- copy ----

// shallowCopy is copy.copy: a new container with the same items, or the
// same object when it is immutable.
func shallowCopy(v any) any {
	v = Unbox(v)
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Pointer || rv.IsNil() {
		return v
	}
	if _, ok := v.(*ClassRef); ok {
		return v
	}
	if _, ok := v.(*FrozenSet); ok {
		return v
	}
	cp := reflect.New(rv.Elem().Type())
	copyFields(cp.Elem(), rv.Elem(), nil)
	return cp.Interface()
}

// deepCopy is copy.deepcopy, with a memo so that shared and cyclic
// references are copied once.
func deepCopy(v any, memo map[uintptr]reflect.Value) any {
	v = Unbox(v)
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return v
	}
	switch rv.Kind() {
	case reflect.Pointer:
		if rv.IsNil() {
			return v
		}
		switch v.(type) {
		case *ClassRef, *FrozenSet, *Callable:
			return v
		}
		if done, ok := memo[rv.Pointer()]; ok {
			return done.Interface()
		}
		cp := reflect.New(rv.Elem().Type())
		memo[rv.Pointer()] = cp
		copyFields(cp.Elem(), rv.Elem(), memo)
		return cp.Interface()
	case reflect.Func, reflect.Chan:
		return v
	}
	out := reflect.New(rv.Type()).Elem()
	copyFields(out, rv, memo)
	return out.Interface()
}

// copyFields sets dst to src. With a memo the values below are copied too.
func copyFields(dst, src reflect.Value, memo map[uintptr]reflect.Value) {
	set := func(d, s reflect.Value) {
		d = reflect.NewAt(d.Type(), unsafe.Pointer(d.UnsafeAddr())).Elem()
		d.Set(s)
	}
	var walk func(d, s reflect.Value)
	walk = func(d, s reflect.Value) {
		switch s.Kind() {
		case reflect.Struct:
			for i := 0; i < s.NumField(); i++ {
				walk(d.Field(i), forceRead(s.Field(i)))
			}
		case reflect.Slice:
			if s.IsNil() {
				return
			}
			n := reflect.MakeSlice(s.Type(), s.Len(), s.Len())
			for i := 0; i < s.Len(); i++ {
				walk(n.Index(i), forceRead(s.Index(i)))
			}
			set(d, n)
		case reflect.Map:
			if s.IsNil() {
				return
			}
			n := reflect.MakeMapWithSize(s.Type(), s.Len())
			for _, k := range s.MapKeys() {
				ev := reflect.New(s.Type().Elem()).Elem()
				walk(ev, forceRead(s.MapIndex(k)))
				n.SetMapIndex(k, ev)
			}
			set(d, n)
		case reflect.Pointer:
			if s.IsNil() || memo == nil {
				set(d, s)
				return
			}
			switch s.Interface().(type) {
			case *ClassRef, *FrozenSet, *Callable:
				set(d, s)
				return
			}
			if done, ok := memo[s.Pointer()]; ok {
				set(d, done)
				return
			}
			cp := reflect.New(s.Elem().Type())
			memo[s.Pointer()] = cp
			walk(cp.Elem(), forceRead(s.Elem()))
			set(d, cp)
		case reflect.Interface:
			if s.IsNil() || memo == nil {
				set(d, s)
				return
			}
			inner := s.Elem()
			cp := reflect.New(inner.Type()).Elem()
			if inner.Kind() == reflect.Pointer {
				cpv := deepCopy(inner.Interface(), memo)
				set(d, reflect.ValueOf(cpv))
				return
			}
			walk(cp, forceRead(inner))
			set(d, cp)
		default:
			set(d, s)
		}
	}
	walk(dst, src)
}

// forceRead makes a value read from an unexported field usable.
func forceRead(v reflect.Value) reflect.Value {
	if v.CanInterface() || !v.CanAddr() {
		return v
	}
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func init() {
	RegisterNative("secrets.token_bytes", func(args []any, kw []KW) any {
		nativeArgs("token_bytes", args, 1, 1)
		n := DynInt(args[0])
		if n < 0 {
			panic(valueError("negative argument not allowed"))
		}
		b := make([]byte, n)
		if _, err := cryptoRand.Read(b); err != nil {
			panic(newError("OSError", err.Error()))
		}
		return Bytes(b)
	})
	RegisterNative("asyncio.isclass", func(args []any, kw []KW) any {
		nativeArgs("isclass", args, 1, 1)
		_, ok := Unbox(args[0]).(*ClassRef)
		return ok
	})
	RegisterNative("asyncio.iscoroutinefunction", func(args []any, kw []KW) any {
		nativeArgs("iscoroutinefunction", args, 1, 1)
		info := funcInfoOf(Unbox(args[0]))
		return info != nil && info.Async
	})
}

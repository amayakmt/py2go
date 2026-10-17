package runtime

import (
	"fmt"
	"strings"
)

// MemoryView is memoryview(obj) over bytes or a bytearray: format 'B', one
// dimension. A view of a bytearray sees and makes changes to it; a slice with
// a step is a copy.
type MemoryView struct {
	arr      *ByteArray // the bytearray viewed, or nil
	data     string     // the bytes viewed when arr is nil
	start    int
	length   int
	released bool
	obj      any
}

// NewMemoryView is memoryview(obj).
func NewMemoryView(args []any, kw []KW) any {
	if len(args) != 1 {
		panic(newError("TypeError", fmt.Sprintf("memoryview() takes exactly 1 argument (%d given)", len(args))))
	}
	switch x := Unbox(args[0]).(type) {
	case *ByteArray:
		return &MemoryView{arr: x, length: len(x.b), obj: x}
	case Bytes:
		return &MemoryView{data: string(x), length: len(x), obj: x}
	case *MemoryView:
		x.check()
		return &MemoryView{arr: x.arr, data: x.data, start: x.start, length: x.length, obj: x.obj}
	}
	if b, ok := baseValueOf(Unbox(args[0])); ok {
		return NewMemoryView([]any{b}, nil)
	}
	panic(newError("TypeError", "memoryview: a bytes-like object is required, not '"+TypeName(Unbox(args[0]))+"'"))
}

func (m *MemoryView) check() {
	if m.released {
		panic(valueError("operation forbidden on released memoryview object"))
	}
}

func (m *MemoryView) bytes() string {
	m.check()
	if m.arr != nil {
		end := m.start + m.length
		if end > len(m.arr.b) {
			end = len(m.arr.b)
		}
		if m.start > end {
			return ""
		}
		return string(m.arr.b[m.start:end])
	}
	return m.data[m.start : m.start+m.length]
}

func (*MemoryView) PyClassName() string { return "memoryview" }
func (m *MemoryView) PyRepr() string {
	if m.released {
		return fmt.Sprintf("<released memory at %#x>", Id(m))
	}
	return fmt.Sprintf("<memory at %#x>", Id(m))
}
func (m *MemoryView) Len() int64 { m.check(); return int64(m.length) }

func (m *MemoryView) IterAny() Cursor[any] {
	s := m.bytes()
	i := 0
	return &funcCursorAny{next: func() (any, bool) {
		if i >= len(s) {
			return nil, false
		}
		i++
		return int64(s[i-1]), true
	}}
}

type funcCursorAny struct {
	next func() (any, bool)
	cur  any
}

func (c *funcCursorAny) Next() bool {
	v, ok := c.next()
	c.cur = v
	return ok
}
func (c *funcCursorAny) Value() any { return c.cur }

func (m *MemoryView) GetItemAny(key any) any {
	m.check()
	key = Unbox(key)
	if sl, ok := key.(Slice); ok {
		bd := sl.Bounds()
		start, step, count := bd.indices(int64(m.length))
		if !bd.HasStep || bd.Step == 1 {
			return &MemoryView{arr: m.arr, data: m.data, start: m.start + int(start), length: int(count), obj: m.obj}
		}
		s := m.bytes()
		out := make([]byte, 0, count)
		for i := int64(0); i < count; i++ {
			out = append(out, s[start+i*step])
		}
		return &MemoryView{data: string(out), length: len(out), obj: m.obj}
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "memoryview: invalid slice key"))
	}
	if i < 0 {
		i += int64(m.length)
	}
	if i < 0 || i >= int64(m.length) {
		panic(indexError("index out of bounds on dimension 1"))
	}
	return int64(m.bytes()[i])
}

func (m *MemoryView) SetItemAny(key, v any) {
	m.check()
	if m.arr == nil {
		panic(newError("TypeError", "cannot modify read-only memory"))
	}
	key = Unbox(key)
	if sl, ok := key.(Slice); ok {
		bd := sl.Bounds()
		start, step, count := bd.indices(int64(m.length))
		repl, isBytes := bytesOf(v)
		if !isBytes {
			panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(Unbox(v))+"'"))
		}
		if int64(len(repl)) != count {
			panic(valueError("memoryview assignment: lvalue and rvalue have different structures"))
		}
		for i := int64(0); i < count; i++ {
			m.arr.b[m.start+int(start+i*step)] = repl[i]
		}
		return
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "memoryview: invalid slice key"))
	}
	if i < 0 {
		i += int64(m.length)
	}
	if i < 0 || i >= int64(m.length) {
		panic(indexError("index out of bounds on dimension 1"))
	}
	m.arr.b[m.start+int(i)] = byteValue(v)
}

func (m *MemoryView) PyEq(other any) (bool, bool) {
	if s, ok := bytesOf(other); ok {
		return m.bytes() == s, true
	}
	return false, true
}

var memoryviewMethodNames = names("tobytes tolist hex release cast toreadonly __enter__ __exit__")

func (*MemoryView) PyDynTable() map[string]bool { return memoryviewMethodNames }

func (m *MemoryView) DynMethod(name string, args []any, kw []KW) (any, bool) {
	switch name {
	case "tobytes":
		return Bytes(m.bytes()), true
	case "tolist":
		s := m.bytes()
		out := NewList[any]()
		for i := 0; i < len(s); i++ {
			out.Append(int64(s[i]))
		}
		return out, true
	case "hex":
		var b strings.Builder
		for _, c := range []byte(m.bytes()) {
			fmt.Fprintf(&b, "%02x", c)
		}
		return b.String(), true
	case "release":
		m.released = true
		return nil, true
	case "cast":
		m.check()
		if len(args) > 0 {
			if f, _ := Unbox(args[0]).(string); f != "B" && f != "b" && f != "c" {
				panic(valueError("memoryview: only the formats 'B', 'b' and 'c' are supported"))
			}
		}
		return &MemoryView{arr: m.arr, data: m.data, start: m.start, length: m.length, obj: m.obj}, true
	case "toreadonly":
		return &MemoryView{data: m.bytes(), length: m.length, obj: m.obj}, true
	case "__enter__":
		return m, true
	case "__exit__":
		m.released = true
		return false, true
	}
	return nil, false
}

func (m *MemoryView) PyGetAttr(name string) (any, bool) {
	switch name {
	case "nbytes":
		return int64(m.length), true
	case "readonly":
		return m.arr == nil, true
	case "obj":
		return m.obj, true
	case "format":
		return "B", true
	case "itemsize", "ndim":
		return int64(1), true
	case "shape":
		return NewVarTuple[any](int64(m.length)), true
	case "strides":
		return NewVarTuple[any](int64(1)), true
	case "released":
		return m.released, true
	case "contiguous", "c_contiguous", "f_contiguous":
		return true, true
	}
	return nil, false
}

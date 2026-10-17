package runtime

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Bytes is a bytes object: immutable, and a Go string underneath so that it
// can be a map key. ByteArray is the mutable one.
type Bytes string

// ByteArray is a bytearray.
type ByteArray struct{ b []byte }

// byteSeq is what bytes and bytearray have in common for the methods that only
// read.
type byteSeq interface {
	~string
}

func (Bytes) PyClassName() string      { return "bytes" }
func (*ByteArray) PyClassName() string { return "bytearray" }

func (b Bytes) PyRepr() string      { return bytesRepr(string(b), "") }
func (b *ByteArray) PyRepr() string { return bytesRepr(string(b.b), "bytearray") }
func (b Bytes) PyStr() string       { return b.PyRepr() }
func (b *ByteArray) PyStr() string  { return b.PyRepr() }
func (b Bytes) PyLen() (int64, bool) {
	return int64(len(b)), true
}
func (b *ByteArray) PyLen() (int64, bool) { return int64(len(b.b)), true }
func (b Bytes) PyTruth() bool             { return len(b) > 0 }
func (b *ByteArray) PyTruth() bool        { return len(b.b) > 0 }
func (b Bytes) Len() int64                { return int64(len(b)) }
func (b *ByteArray) Len() int64           { return int64(len(b.b)) }
func (b Bytes) PyHash() int64             { return fixHash(int64(siphash13(hashK0, hashK1, []byte(b)))) }
func (b *ByteArray) PyHash() int64 {
	panic(newError("TypeError", "unhashable type: 'bytearray'"))
}

// bytesRepr is repr(b"...") and repr(bytearray(b"...")).
func bytesRepr(s, wrap string) string {
	q := byte('\'')
	if strings.IndexByte(s, '\'') >= 0 && strings.IndexByte(s, '"') < 0 {
		q = '"'
	}
	var b strings.Builder
	if wrap != "" {
		b.WriteString(wrap + "(")
	}
	b.WriteByte('b')
	b.WriteByte(q)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == q || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c == '\n':
			b.WriteString(`\n`)
		case c == '\r':
			b.WriteString(`\r`)
		case c == '\t':
			b.WriteString(`\t`)
		case c < 0x20 || c >= 0x7f:
			fmt.Fprintf(&b, `\x%02x`, c)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte(q)
	if wrap != "" {
		b.WriteByte(')')
	}
	return b.String()
}

// bytesOf reads the bytes of a bytes-like object.
func bytesOf(v any) (string, bool) {
	switch x := BaseValue(Unbox(v)).(type) {
	case *MemoryView:
		return x.bytes(), true
	case Bytes:
		return string(x), true
	case *ByteArray:
		return string(x.b), true
	}
	return "", false
}

func needBytes(v any, what string) string {
	s, ok := bytesOf(v)
	if !ok {
		panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(Unbox(v))+"'"))
	}
	_ = what
	return s
}

// ---- construction ----

// NewBytes is bytes(x...): bytes(), bytes(n), bytes(iterable of ints),
// bytes(text, encoding[, errors]) and bytes(bytes-like).
func NewBytes(args []any, kw []KW) Bytes {
	return Bytes(buildBytes("bytes", args, kw))
}

// NewByteArray is bytearray(x...).
func NewByteArray(args []any, kw []KW) *ByteArray {
	return &ByteArray{b: []byte(buildBytes("bytearray", args, kw))}
}

func buildBytes(name string, args []any, kw []KW) string {
	if len(args) == 0 {
		if _, ok := kwValue(kw, "source"); !ok {
			return ""
		}
	}
	var src any
	switch {
	case len(args) > 0:
		src = Unbox(args[0])
	default:
		src, _ = kwValue(kw, "source")
	}
	enc, hasEnc := "utf-8", false
	if len(args) > 1 {
		enc, hasEnc = argStr(args, 1, "utf-8"), true
	} else if v, ok := kwValue(kw, "encoding"); ok {
		enc, hasEnc = Unbox(v).(string), true
	}
	errs := argStr(args, 2, "strict")
	if v, ok := kwValue(kw, "errors"); ok {
		errs = Unbox(v).(string)
	}
	switch x := src.(type) {
	case string:
		if !hasEnc {
			panic(newError("TypeError", "string argument without an encoding"))
		}
		return string(EncodeStr(x, enc, errs))
	case int64, bool, *big.Int:
		if hasEnc {
			panic(newError("TypeError", "encoding without a string argument"))
		}
		n, ok := intOf(x)
		if !ok {
			panic(overflowError("cannot fit 'int' into an index-sized integer"))
		}
		if n < 0 {
			panic(valueError("negative count"))
		}
		return strings.Repeat("\x00", int(n))
	case Bytes:
		return string(x)
	case *ByteArray:
		return string(x.b)
	case float64:
		panic(newError("TypeError", "cannot convert 'float' object to "+name))
	}
	if hasEnc {
		panic(newError("TypeError", "encoding without a string argument"))
	}
	if _, named := src.(Named); named && DynHasAttr(src, "__bytes__") {
		r := DynCall(DynGetAttr(src, "__bytes__"))
		if bs, ok := r.(Bytes); ok {
			return string(bs)
		}
		panic(newError("TypeError", "__bytes__ returned non-bytes (type "+TypeName(r)+")"))
	}
	var out []byte
	for c := DynIter(src).Iter(); c.Next(); {
		n, ok := intOf(Unbox(c.Value()))
		if !ok {
			panic(newError("TypeError", "'"+TypeName(Unbox(c.Value()))+"' object cannot be interpreted as an integer"))
		}
		if n < 0 || n > 255 {
			panic(valueError("bytes must be in range(0, 256)"))
		}
		out = append(out, byte(n))
	}
	return string(out)
}

// ---- protocol hooks ----

func (b Bytes) IterAny() Cursor[any] { return &byteCursor{s: string(b)} }
func (b *ByteArray) IterAny() Cursor[any] {
	return &byteCursor{s: string(b.b)}
}

type byteCursor struct {
	s   string
	i   int
	cur any
}

func (c *byteCursor) Next() bool {
	if c.i >= len(c.s) {
		return false
	}
	c.cur = int64(c.s[c.i])
	c.i++
	return true
}
func (c *byteCursor) Value() any { return c.cur }

func (b Bytes) ContainsAny(item any) bool      { return bytesContains(string(b), item) }
func (b *ByteArray) ContainsAny(item any) bool { return bytesContains(string(b.b), item) }

func bytesContains(s string, item any) bool {
	item = Unbox(item)
	if n, ok := intOf(item); ok && !isBig(item) {
		if n < 0 || n > 255 {
			panic(valueError("byte must be in range(0, 256)"))
		}
		return strings.IndexByte(s, byte(n)) >= 0
	}
	sub, ok := bytesOf(item)
	if !ok {
		panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(item)+"'"))
	}
	return strings.Contains(s, sub)
}

func (b Bytes) GetItemAny(key any) any      { return bytesGet(string(b), key, false) }
func (b *ByteArray) GetItemAny(key any) any { return bytesGet(string(b.b), key, true) }

func bytesGet(s string, key any, mutable bool) any {
	key = Unbox(key)
	if sl, ok := key.(Slice); ok {
		start, step, count := sl.Bounds().indices(int64(len(s)))
		out := make([]byte, 0, count)
		for i := int64(0); i < count; i++ {
			out = append(out, s[start+i*step])
		}
		if mutable {
			return &ByteArray{b: out}
		}
		return Bytes(out)
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "byte indices must be integers or slices, not "+TypeName(key)))
	}
	if i < 0 {
		i += int64(len(s))
	}
	if i < 0 || i >= int64(len(s)) {
		if mutable {
			panic(indexError("bytearray index out of range"))
		}
		panic(indexError("index out of range"))
	}
	return int64(s[i])
}

func (b *ByteArray) SetItemAny(key, v any) {
	key = Unbox(key)
	if sl, ok := key.(Slice); ok {
		repl, isBytes := bytesOf(v)
		if !isBytes {
			var tmp []byte
			for c := DynIter(v).Iter(); c.Next(); {
				n, _ := intOf(Unbox(c.Value()))
				tmp = append(tmp, byte(n))
			}
			repl = string(tmp)
		}
		bd := sl.Bounds()
		start, step, count := bd.indices(int64(len(b.b)))
		if !bd.HasStep || bd.Step == 1 {
			end := start + count
			out := append([]byte{}, b.b[:start]...)
			out = append(out, repl...)
			out = append(out, b.b[end:]...)
			b.b = out
			return
		}
		if int64(len(repl)) != count {
			panic(valueError(fmt.Sprintf("attempt to assign bytes of size %d to extended slice of size %d", len(repl), count)))
		}
		for i := int64(0); i < count; i++ {
			b.b[start+i*step] = repl[i]
		}
		return
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "bytearray indices must be integers or slices, not "+TypeName(key)))
	}
	nb := byteValue(v)
	n := int64(nb)
	if i < 0 {
		i += int64(len(b.b))
	}
	if i < 0 || i >= int64(len(b.b)) {
		panic(indexError("bytearray index out of range"))
	}
	b.b[i] = byte(n)
}

func (b *ByteArray) DelItemAny(key any) {
	key = Unbox(key)
	if sl, ok := key.(Slice); ok {
		start, step, count := sl.Bounds().indices(int64(len(b.b)))
		drop := map[int64]bool{}
		for i := int64(0); i < count; i++ {
			drop[start+i*step] = true
		}
		out := b.b[:0:0]
		for i, c := range b.b {
			if !drop[int64(i)] {
				out = append(out, c)
			}
		}
		b.b = out
		return
	}
	i, ok := indexArg(key)
	if !ok {
		panic(newError("TypeError", "bytearray indices must be integers or slices, not "+TypeName(key)))
	}
	if i < 0 {
		i += int64(len(b.b))
	}
	if i < 0 || i >= int64(len(b.b)) {
		panic(indexError("bytearray index out of range"))
	}
	b.b = append(b.b[:i], b.b[i+1:]...)
}

// ---- comparison ----

func (b Bytes) PyEq(other any) (bool, bool) {
	o, ok := bytesOf(other)
	if !ok {
		return false, true
	}
	return string(b) == o, true
}
func (b *ByteArray) PyEq(other any) (bool, bool) {
	o, ok := bytesOf(other)
	if !ok {
		return false, true
	}
	return string(b.b) == o, true
}

func (b Bytes) PyOrder(op string, other any) (bool, bool) { return bytesOrder(string(b), op, other) }
func (b *ByteArray) PyOrder(op string, other any) (bool, bool) {
	return bytesOrder(string(b.b), op, other)
}

func bytesOrder(s, op string, other any) (bool, bool) {
	o, ok := bytesOf(other)
	if !ok {
		return false, false
	}
	switch op {
	case "lt":
		return s < o, true
	case "le":
		return s <= o, true
	case "gt":
		return s > o, true
	}
	return s >= o, true
}

// ---- operators ----

func (b Bytes) PyBinop(op string, o any, reflected bool) (any, bool) {
	return bytesBinop(string(b), false, op, o, reflected)
}
func (b *ByteArray) PyBinop(op string, o any, reflected bool) (any, bool) {
	return bytesBinop(string(b.b), true, op, o, reflected)
}

func bytesBinop(s string, mutable bool, op string, o any, reflected bool) (any, bool) {
	make := func(x string) any {
		if mutable {
			return &ByteArray{b: []byte(x)}
		}
		return Bytes(x)
	}
	switch op {
	case "Add":
		other, ok := bytesOf(o)
		if !ok {
			if reflected {
				return nil, false
			}
			panic(newError("TypeError", "can't concat "+TypeName(Unbox(o))+" to "+map[bool]string{true: "bytearray", false: "bytes"}[mutable]))
		}
		if reflected {
			return make(other + s), true
		}
		return make(s + other), true
	case "Mult":
		n, ok := intOf(Unbox(o))
		if !ok {
			return nil, false
		}
		if n < 0 {
			n = 0
		}
		return make(strings.Repeat(s, int(n))), true
	case "Mod":
		if !reflected {
			return make(bytesPercent(s, o)), true
		}
	}
	return nil, false
}

// bytesPercent is b"..." % args: the % codes of str, with %b and %a added.
// The conversions that hold bytes are handled here; the numeric ones go
// through the str formatter one at a time.
func bytesPercent(format string, args any) string {
	args = Unbox(args)
	var list []any
	tuple := false
	if t, ok := args.(Seq); ok && t.Kind() == "tuple" {
		tuple = true
		for i := int64(0); i < t.Len(); i++ {
			list = append(list, t.At(i))
		}
	} else {
		list = []any{args}
	}
	mapping := false
	if _, ok := args.(Seq); !ok {
		if _, isStr := args.(string); !isStr {
			if _, isB := bytesOf(args); !isB {
				_, mapping = args.(interface{ Len() int64 })
			}
		}
	}
	next := 0
	take := func() any {
		if next >= len(list) {
			panic(newError("TypeError", "not enough arguments for format string"))
		}
		next++
		return list[next-1]
	}
	var out strings.Builder
	usedKey := false
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' {
			out.WriteByte(c)
			continue
		}
		i++
		var arg any
		haveArg := false
		if i < len(format) && format[i] == '(' {
			j := strings.IndexByte(format[i:], ')')
			if j < 0 {
				panic(valueError("incomplete format key"))
			}
			arg = DynGetItem(args, Bytes(format[i+1:i+j]))
			haveArg, usedKey = true, true
			i += j + 1
		}
		flags := ""
		for i < len(format) && strings.IndexByte("-+ #0", format[i]) >= 0 {
			flags += string(format[i])
			i++
		}
		width, prec := "", ""
		if i < len(format) && format[i] == '*' {
			width = strconv.FormatInt(DynInt(take()), 10)
			i++
		}
		for i < len(format) && format[i] >= '0' && format[i] <= '9' {
			width += string(format[i])
			i++
		}
		if i < len(format) && format[i] == '.' {
			prec = "."
			i++
			if i < len(format) && format[i] == '*' {
				prec += strconv.FormatInt(DynInt(take()), 10)
				i++
			}
			for i < len(format) && format[i] >= '0' && format[i] <= '9' {
				prec += string(format[i])
				i++
			}
		}
		for i < len(format) && strings.IndexByte("hlL", format[i]) >= 0 {
			i++
		}
		if i >= len(format) {
			panic(valueError("incomplete format"))
		}
		verb := format[i]
		if verb == '%' {
			out.WriteByte('%')
			continue
		}
		if !haveArg {
			if strings.IndexByte("bsraciduoxXeEfFgG", verb) < 0 {
				panic(valueError(fmt.Sprintf("unsupported format character '%c' (0x%x) at index %d", verb, verb, i)))
			}
			arg = take()
		}
		var text string
		switch verb {
		case 'b', 's':
			if bs, ok := bytesOf(arg); ok {
				text = bs
			} else {
				panic(newError("TypeError", "%b requires a bytes-like object, or an object that implements __bytes__, not '"+TypeName(Unbox(arg))+"'"))
			}
			if prec != "" {
				n, _ := strconv.Atoi(prec[1:])
				if n < len(text) {
					text = text[:n]
				}
			}
		case 'r', 'a':
			text = Ascii(arg)
			if prec != "" {
				n, _ := strconv.Atoi(prec[1:])
				if n < len(text) {
					text = text[:n]
				}
			}
		case 'c':
			if bs, ok := bytesOf(arg); ok {
				if len(bs) != 1 {
					panic(newError("TypeError", "%c requires an integer in range(256) or a single byte"))
				}
				text = bs
			} else if n, ok := intOf(Unbox(arg)); ok && !isBig(Unbox(arg)) {
				if n < 0 || n > 255 {
					panic(overflowError("%c arg not in range(256)"))
				}
				text = string([]byte{byte(n)})
			} else {
				panic(newError("TypeError", "%c requires an integer in range(256) or a single byte"))
			}
		default:
			if strings.IndexByte("iduoxXeEfFgG", verb) < 0 {
				panic(valueError(fmt.Sprintf("unsupported format character '%c' (0x%x) at index %d", verb, verb, i)))
			}
			out.WriteString(PercentFormat("%"+flags+width+prec+string(verb), NewVarTuple(arg)))
			continue
		}
		if n, _ := strconv.Atoi(width); n > len(text) {
			pad := strings.Repeat(" ", n-len(text))
			if strings.Contains(flags, "-") {
				text += pad
			} else {
				text = pad + text
			}
		}
		out.WriteString(text)
	}
	if next < len(list) && !usedKey && !mapping {
		panic(newError("TypeError", "not all arguments converted during bytes formatting"))
	}
	_ = tuple
	return out.String()
}

func bytesToText(s string) string {
	r := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		r[i] = rune(s[i])
	}
	return string(r)
}

func textToBytes(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		out = append(out, byte(r))
	}
	return string(out)
}

func bytesFormatArgs(args any) any {
	args = Unbox(args)
	conv := func(v any) any {
		if s, ok := bytesOf(v); ok {
			return bytesToText(s)
		}
		return v
	}
	if t, ok := args.(Seq); ok && t.Kind() == "tuple" {
		items := make([]any, t.Len())
		for i := range items {
			items[i] = conv(t.At(int64(i)))
		}
		return NewVarTuple(items...)
	}
	return conv(args)
}

// ---- methods ----

var bytesMethodNames = names("decode hex startswith endswith find rfind index rindex count replace split rsplit join strip lstrip rstrip upper lower swapcase capitalize title isalpha isdigit isalnum isspace isupper islower istitle isascii partition rpartition center ljust rjust zfill splitlines expandtabs translate removeprefix removesuffix")
var byteArrayMethodNames = names("decode hex startswith endswith find rfind index rindex count replace split rsplit join strip lstrip rstrip upper lower swapcase capitalize title isalpha isdigit isalnum isspace isupper islower istitle isascii partition rpartition center ljust rjust zfill splitlines expandtabs translate removeprefix removesuffix append extend pop insert remove reverse clear copy")

func (Bytes) PyDynTable() map[string]bool      { return bytesMethodNames }
func (*ByteArray) PyDynTable() map[string]bool { return byteArrayMethodNames }

func (b Bytes) DynMethod(name string, args []any, kw []KW) (any, bool) {
	return bytesMethod(string(b), nil, name, args, kw)
}
func (b *ByteArray) DynMethod(name string, args []any, kw []KW) (any, bool) {
	return bytesMethod(string(b.b), b, name, args, kw)
}

func bytesMethod(s string, ba *ByteArray, name string, args []any, kw []KW) (any, bool) {
	result := func(x string) any {
		if ba != nil {
			return &ByteArray{b: []byte(x)}
		}
		return Bytes(x)
	}
	byteArg := func(i int) string { return needBytes(argAt(args, i), name) }
	switch name {
	case "decode":
		enc, errs := argNamed("decode", "encoding", args, 0, "utf-8"), argNamed("decode", "errors", args, 1, "strict")
		if v, ok := kwValue(kw, "encoding"); ok {
			enc = Unbox(v).(string)
		}
		if v, ok := kwValue(kw, "errors"); ok {
			errs = Unbox(v).(string)
		}
		return DecodeBytes(s, enc, errs), true
	case "hex":
		sep := ""
		if len(args) > 0 {
			sep = argStr(args, 0, "")
		}
		per := int(argInt(args, 1, 1))
		return hexBytes(s, sep, per), true
	case "startswith", "endswith":
		var prefixes []string
		switch p := Unbox(argAt(args, 0)).(type) {
		case Seq:
			for i := int64(0); i < p.Len(); i++ {
				prefixes = append(prefixes, needBytes(p.At(i), name))
			}
		default:
			bs, ok := bytesOf(p)
			if !ok {
				panic(newError("TypeError", name+" first arg must be bytes or a tuple of bytes, not "+TypeName(p)))
			}
			prefixes = []string{bs}
		}
		start, end, found := searchRange(len(s), args, 1)
		if !found {
			return false, true
		}
		hay := s[start:end]
		for _, p := range prefixes {
			if (name == "startswith" && strings.HasPrefix(hay, p)) || (name == "endswith" && strings.HasSuffix(hay, p)) {
				return true, true
			}
		}
		return false, true
	case "find", "rfind", "index", "rindex", "count":
		var sub string
		if n, ok := intOf(Unbox(argAt(args, 0))); ok && !isBig(Unbox(argAt(args, 0))) {
			if n < 0 || n > 255 {
				panic(valueError("byte must be in range(0, 256)"))
			}
			sub = string([]byte{byte(n)})
		} else if s2, ok := bytesOf(argAt(args, 0)); ok {
			sub = s2
		} else {
			panic(newError("TypeError", "argument should be integer or bytes-like object, not '"+TypeName(Unbox(argAt(args, 0)))+"'"))
		}
		start, end, found := searchRange(len(s), args, 1)
		if !found {
			if name == "count" {
				return int64(0), true
			}
			if name == "index" || name == "rindex" {
				panic(valueError("subsection not found"))
			}
			return int64(-1), true
		}
		hay := s[start:end]
		switch name {
		case "count":
			if sub == "" {
				return int64(len(hay) + 1), true
			}
			return int64(strings.Count(hay, sub)), true
		case "find", "index":
			at := strings.Index(hay, sub)
			if at < 0 {
				if name == "index" {
					panic(valueError("subsection not found"))
				}
				return int64(-1), true
			}
			return int64(start + at), true
		}
		at := strings.LastIndex(hay, sub)
		if at < 0 {
			if name == "rindex" {
				panic(valueError("subsection not found"))
			}
			return int64(-1), true
		}
		return int64(start + at), true
	case "replace":
		n := int(argInt(args, 2, -1))
		return result(strings.Replace(s, byteArg(0), byteArg(1), n)), true
	case "split", "rsplit":
		var parts []string
		max := argInt(args, 1, -1)
		if v, ok := kwValue(kw, "maxsplit"); ok {
			max, _ = intOf(Unbox(v))
		}
		sepArg := argAt(args, 0)
		if v, ok := kwValue(kw, "sep"); ok {
			sepArg = v
		}
		if IsNone(sepArg) {
			parts = splitFields(s, max, name == "rsplit", func(c byte) bool { return isByteSpace(c) })
		} else {
			sep := needBytes(sepArg, name)
			if sep == "" {
				panic(valueError("empty separator"))
			}
			parts = splitBy(s, sep, max, name == "rsplit")
		}
		out := NewList[any]()
		for _, p := range parts {
			out.Append(result(p))
		}
		return out, true
	case "join":
		var parts []string
		var src Iterable[any]
		if e := Catch(func() { src = DynIter(argAt(args, 0)) }); e != nil {
			panic(newError("TypeError", "can only join an iterable"))
		}
		for c := src.Iter(); c.Next(); {
			p, ok := bytesOf(c.Value())
			if !ok {
				panic(newError("TypeError", fmt.Sprintf("sequence item %d: expected a bytes-like object, %s found", len(parts), TypeName(Unbox(c.Value())))))
			}
			parts = append(parts, p)
		}
		return result(strings.Join(parts, s)), true
	case "strip", "lstrip", "rstrip":
		set := " \t\n\r\x0b\x0c"
		if len(args) > 0 && !IsNone(args[0]) {
			set = byteArg(0)
		}
		switch name {
		case "strip":
			return result(strings.Trim(s, set)), true
		case "lstrip":
			return result(strings.TrimLeft(s, set)), true
		}
		return result(strings.TrimRight(s, set)), true
	case "upper":
		return result(asciiMap(s, 'a', 'z', -32)), true
	case "lower":
		return result(asciiMap(s, 'A', 'Z', 32)), true
	case "swapcase":
		b := []byte(s)
		for i, c := range b {
			switch {
			case c >= 'a' && c <= 'z':
				b[i] = c - 32
			case c >= 'A' && c <= 'Z':
				b[i] = c + 32
			}
		}
		return result(string(b)), true
	case "capitalize":
		if s == "" {
			return result(s), true
		}
		return result(asciiMap(s[:1], 'a', 'z', -32) + asciiMap(s[1:], 'A', 'Z', 32)), true
	case "title":
		b := []byte(s)
		prev := false
		for i, c := range b {
			letter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
			switch {
			case letter && !prev && c >= 'a':
				b[i] = c - 32
			case letter && prev && c <= 'Z':
				b[i] = c + 32
			}
			prev = letter
		}
		return result(string(b)), true
	case "isalpha", "isdigit", "isalnum", "isspace", "isupper", "islower", "isascii", "istitle":
		return bytesPredicate(s, name), true
	case "partition", "rpartition":
		sep := byteArg(0)
		if sep == "" {
			panic(valueError("empty separator"))
		}
		var i int
		if name == "partition" {
			i = strings.Index(s, sep)
		} else {
			i = strings.LastIndex(s, sep)
		}
		switch {
		case i >= 0:
			return NewVarTuple[any](result(s[:i]), result(sep), result(s[i+len(sep):])), true
		case name == "partition":
			return NewVarTuple[any](result(s), result(""), result("")), true
		}
		return NewVarTuple[any](result(""), result(""), result(s)), true
	case "center", "ljust", "rjust":
		w := int(argInt(args, 0, 0))
		fill := " "
		if len(args) > 1 {
			fill = byteArg(1)
		}
		if len(s) >= w {
			return result(s), true
		}
		pad := w - len(s)
		switch name {
		case "ljust":
			return result(s + strings.Repeat(fill, pad)), true
		case "rjust":
			return result(strings.Repeat(fill, pad) + s), true
		}
		left := pad/2 + (pad & w & 1)
		return result(strings.Repeat(fill, left) + s + strings.Repeat(fill, pad-left)), true
	case "zfill":
		w := int(argInt(args, 0, 0))
		if len(s) >= w {
			return result(s), true
		}
		sign, body := "", s
		if s != "" && (s[0] == '-' || s[0] == '+') {
			sign, body = s[:1], s[1:]
		}
		return result(sign + strings.Repeat("0", w-len(s)) + body), true
	case "splitlines":
		keep := false
		if len(args) > 0 {
			keep = Truth(Unbox(args[0]))
		}
		if v, ok := kwValue(kw, "keepends"); ok {
			keep = Truth(Unbox(v))
		}
		out := NewList[any]()
		for len(s) > 0 {
			i := strings.IndexAny(s, "\n\r")
			if i < 0 {
				out.Append(result(s))
				break
			}
			j := i + 1
			if s[i] == '\r' && j < len(s) && s[j] == '\n' {
				j++
			}
			if keep {
				out.Append(result(s[:j]))
			} else {
				out.Append(result(s[:i]))
			}
			s = s[j:]
		}
		return out, true
	case "expandtabs":
		return result(strExpandTabs(s, int(argInt(args, 0, 8)))), true
	case "translate":
		table := Unbox(argAt(args, 0))
		del := ""
		if len(args) > 1 {
			del = byteArg(1)
		} else if v, ok := kwValue(kw, "delete"); ok {
			del = needBytes(v, name)
		}
		var tbl string
		if !IsNone(table) {
			tbl = needBytes(table, name)
			if len(tbl) != 256 {
				panic(valueError("translation table must be 256 characters long"))
			}
		}
		out := make([]byte, 0, len(s))
		for i := 0; i < len(s); i++ {
			if strings.IndexByte(del, s[i]) >= 0 {
				continue
			}
			if tbl != "" {
				out = append(out, tbl[s[i]])
			} else {
				out = append(out, s[i])
			}
		}
		return result(string(out)), true
	case "removeprefix":
		return result(strings.TrimPrefix(s, byteArg(0))), true
	case "removesuffix":
		return result(strings.TrimSuffix(s, byteArg(0))), true
	}
	if ba == nil {
		return nil, false
	}
	switch name {
	case "append":
		ba.b = append(ba.b, byteValue(argAt(args, 0)))
		return nil, true
	case "extend":
		for c := DynIter(argAt(args, 0)).Iter(); c.Next(); {
			ba.b = append(ba.b, byteValue(c.Value()))
		}
		return nil, true
	case "pop":
		if len(ba.b) == 0 {
			panic(indexError("pop from empty bytearray"))
		}
		i := int(argInt(args, 0, -1))
		if i < 0 {
			i += len(ba.b)
		}
		if i < 0 || i >= len(ba.b) {
			panic(indexError("pop index out of range"))
		}
		v := ba.b[i]
		ba.b = append(ba.b[:i], ba.b[i+1:]...)
		return int64(v), true
	case "insert":
		i := int(argInt(args, 0, 0))
		v := byteValue(argAt(args, 1))
		if i < 0 {
			i = max(i+len(ba.b), 0)
		}
		i = min(i, len(ba.b))
		ba.b = append(ba.b[:i], append([]byte{v}, ba.b[i:]...)...)
		return nil, true
	case "remove":
		i := bytes.IndexByte(ba.b, byteValue(argAt(args, 0)))
		if i < 0 {
			panic(valueError("value not found in bytearray"))
		}
		ba.b = append(ba.b[:i], ba.b[i+1:]...)
		return nil, true
	case "reverse":
		for i, j := 0, len(ba.b)-1; i < j; i, j = i+1, j-1 {
			ba.b[i], ba.b[j] = ba.b[j], ba.b[i]
		}
		return nil, true
	case "clear":
		ba.b = nil
		return nil, true
	case "copy":
		return &ByteArray{b: append([]byte{}, ba.b...)}, true
	}
	return nil, false
}

func isByteSpace(c byte) bool { return c == ' ' || (c >= 9 && c <= 13) }

func asciiMap(s string, lo, hi byte, delta int) string {
	b := []byte(s)
	for i, c := range b {
		if c >= lo && c <= hi {
			b[i] = byte(int(c) + delta)
		}
	}
	return string(b)
}

func bytesPredicate(s, name string) bool {
	if name == "isascii" {
		for i := 0; i < len(s); i++ {
			if s[i] >= 0x80 {
				return false
			}
		}
		return true
	}
	if s == "" {
		return false
	}
	switch name {
	case "isupper", "islower":
		cased := false
		for i := 0; i < len(s); i++ {
			c := s[i]
			switch {
			case c >= 'A' && c <= 'Z':
				if name == "islower" {
					return false
				}
				cased = true
			case c >= 'a' && c <= 'z':
				if name == "isupper" {
					return false
				}
				cased = true
			}
		}
		return cased
	case "istitle":
		return string(asciiTitle(s)) == s && s != strings.ToLower(s)
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		var ok bool
		switch name {
		case "isalpha":
			ok = (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		case "isdigit":
			ok = c >= '0' && c <= '9'
		case "isalnum":
			ok = (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		case "isspace":
			ok = isByteSpace(c)
		}
		if !ok {
			return false
		}
	}
	return true
}

func asciiTitle(s string) []byte {
	b := []byte(s)
	prev := false
	for i, c := range b {
		letter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		switch {
		case letter && !prev && c >= 'a':
			b[i] = c - 32
		case letter && prev && c <= 'Z':
			b[i] = c + 32
		}
		prev = letter
	}
	return b
}

func hexBytes(s, sep string, per int) string {
	const digits = "0123456789abcdef"
	var b strings.Builder
	n := len(s)
	for i := 0; i < n; i++ {
		if sep != "" && i > 0 && per != 0 {
			if per > 0 && (n-i)%per == 0 || per < 0 && i%(-per) == 0 {
				b.WriteString(sep)
			}
		}
		b.WriteByte(digits[s[i]>>4])
		b.WriteByte(digits[s[i]&15])
	}
	return b.String()
}

// FromHex is bytes.fromhex(text): the bytes, and on a bad digit the position
// of the first one.
func FromHex(text string) (string, int, bool) {
	var out []byte
	runes := []byte(text)
	for i := 0; i < len(runes); {
		c := runes[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f' {
			i++
			continue
		}
		hi, ok1 := hexVal(c)
		if !ok1 {
			return "", i, false
		}
		if i+1 >= len(runes) {
			return "", i + 1, false
		}
		lo, ok2 := hexVal(runes[i+1])
		if !ok2 {
			return "", i + 1, false
		}
		out = append(out, hi<<4|lo)
		i += 2
	}
	return string(out), 0, true
}

func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// splitFields is split() without a separator on bytes: runs of whitespace.
func splitFields(s string, max int64, right bool, isSep func(byte) bool) []string {
	var parts []string
	if !right {
		i := 0
		for i < len(s) {
			for i < len(s) && isSep(s[i]) {
				i++
			}
			if i >= len(s) {
				break
			}
			if max >= 0 && int64(len(parts)) == max {
				parts = append(parts, s[i:])
				return parts
			}
			j := i
			for j < len(s) && !isSep(s[j]) {
				j++
			}
			parts = append(parts, s[i:j])
			i = j
		}
		return parts
	}
	i := len(s)
	for i > 0 {
		for i > 0 && isSep(s[i-1]) {
			i--
		}
		if i == 0 {
			break
		}
		if max >= 0 && int64(len(parts)) == max {
			parts = append(parts, s[:i])
			break
		}
		j := i
		for j > 0 && !isSep(s[j-1]) {
			j--
		}
		parts = append(parts, s[j:i])
		i = j
	}
	for l, r := 0, len(parts)-1; l < r; l, r = l+1, r-1 {
		parts[l], parts[r] = parts[r], parts[l]
	}
	return parts
}

func splitBy(s, sep string, max int64, right bool) []string {
	if max < 0 {
		return strings.Split(s, sep)
	}
	if !right {
		return strings.SplitN(s, sep, int(max)+1)
	}
	var parts []string
	rest := s
	for i := int64(0); i < max; i++ {
		j := strings.LastIndex(rest, sep)
		if j < 0 {
			break
		}
		parts = append([]string{rest[j+len(sep):]}, parts...)
		rest = rest[:j]
	}
	return append([]string{rest}, parts...)
}

// ---- int.to_bytes and int.from_bytes ----

func intToBytes(b *big.Int, args []any, kw []KW) any {
	length := int(argInt(args, 0, 1))
	if v, ok := kwValue(kw, "length"); ok {
		n, _ := intOf(Unbox(v))
		length = int(n)
	}
	if length < 0 {
		panic(valueError("length argument must be non-negative"))
	}
	order := argStr(args, 1, "big")
	if v, ok := kwValue(kw, "byteorder"); ok {
		order = Unbox(v).(string)
	}
	signed := false
	if len(args) > 2 {
		signed = Truth(Unbox(args[2]))
	}
	if v, ok := kwValue(kw, "signed"); ok {
		signed = Truth(Unbox(v))
	}
	if order != "little" && order != "big" {
		panic(valueError("byteorder must be either 'little' or 'big'"))
	}
	if b.Sign() < 0 && !signed {
		panic(overflowError("can't convert negative int to unsigned"))
	}
	m := new(big.Int).Lsh(bigOne, uint(length*8))
	v := new(big.Int).Set(b)
	if signed {
		half := new(big.Int).Lsh(bigOne, uint(length*8-1))
		if v.Cmp(new(big.Int).Neg(half)) < 0 || v.Cmp(half) >= 0 {
			panic(overflowError("int too big to convert"))
		}
		if v.Sign() < 0 {
			v.Add(v, m)
		}
	} else if v.Cmp(m) >= 0 {
		panic(overflowError("int too big to convert"))
	}
	buf := v.FillBytes(make([]byte, length))
	if order == "little" {
		for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
			buf[i], buf[j] = buf[j], buf[i]
		}
	} else if order != "big" {
		panic(valueError("byteorder must be either 'little' or 'big'"))
	}
	return Bytes(buf)
}

// IntFromBytes is int.from_bytes(data, byteorder, signed=False).
func IntFromBytes(args []any, kw []KW) any {
	data := []byte(needBytesIter(argAt(args, 0)))
	order := argStr(args, 1, "big")
	if v, ok := kwValue(kw, "byteorder"); ok {
		order = Unbox(v).(string)
	}
	signed := false
	if v, ok := kwValue(kw, "signed"); ok {
		signed = Truth(Unbox(v))
	}
	if order != "little" && order != "big" {
		panic(valueError("byteorder must be either 'little' or 'big'"))
	}
	if order == "little" {
		for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
			data[i], data[j] = data[j], data[i]
		}
	}
	v := new(big.Int).SetBytes(data)
	if signed && len(data) > 0 && data[0]&0x80 != 0 {
		v.Sub(v, new(big.Int).Lsh(bigOne, uint(len(data)*8)))
	}
	return NormBig(v)
}

func needBytesIter(v any) string {
	if s, ok := bytesOf(v); ok {
		return s
	}
	if _, isStr := Unbox(v).(string); isStr {
		panic(newError("TypeError", "cannot convert 'str' object to bytes"))
	}
	return buildBytes("bytes", []any{v}, nil)
}

var _ = utf8.RuneError

// byteValue reads an int in range(0, 256).
func byteValue(v any) byte {
	v = Unbox(v)
	n, ok := intOf(v)
	if !ok || isBig(v) {
		if isBig(v) {
			panic(valueError("byte must be in range(0, 256)"))
		}
		panic(newError("TypeError", "'"+TypeName(v)+"' object cannot be interpreted as an integer"))
	}
	if n < 0 || n > 255 {
		panic(valueError("byte must be in range(0, 256)"))
	}
	return byte(n)
}

package runtime

import (
	"encoding/binary"
	"math"
	"math/big"
	"strconv"
)

// struct: packing and unpacking of C structures.

const structErr = "struct_pm_error"

func structError(msg string) *PyError { return newError(structErr, msg) }

type structItem struct {
	code   byte
	count  int // repeat count, or the length of s and p
	size   int
	offset int // from the start of the buffer
}

type structFmt struct {
	items  []structItem
	size   int
	little bool
	native bool
	nvals  int // values packed or unpacked
}

var structFmtCache = map[string]*structFmt{}

func nativeSize(c byte) (size, align int, ok bool) {
	switch c {
	case 'x', 'c', 'b', 'B', '?', 's', 'p':
		return 1, 1, true
	case 'h', 'H', 'e':
		return 2, 2, true
	case 'i', 'I', 'f':
		return 4, 4, true
	case 'l', 'L', 'q', 'Q', 'n', 'N', 'd', 'P':
		return 8, 8, true
	}
	return 0, 0, false
}

func stdSize(c byte) (int, bool) {
	switch c {
	case 'x', 'c', 'b', 'B', '?', 's', 'p':
		return 1, true
	case 'h', 'H', 'e':
		return 2, true
	case 'i', 'I', 'l', 'L', 'f':
		return 4, true
	case 'q', 'Q', 'd':
		return 8, true
	}
	return 0, false
}

func parseStructFmt(f string) *structFmt {
	if p, ok := structFmtCache[f]; ok {
		return p
	}
	sf := &structFmt{little: true, native: true}
	i := 0
	if len(f) > 0 {
		switch f[0] {
		case '@':
			i = 1
		case '=':
			sf.native = false
			i = 1
		case '<':
			sf.native, i = false, 1
		case '>', '!':
			sf.native, sf.little, i = false, false, 1
		}
	}
	off := 0
	for i < len(f) {
		c := f[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f' {
			i++
			continue
		}
		count, hasCount := 1, false
		if c >= '0' && c <= '9' {
			n := 0
			for i < len(f) && f[i] >= '0' && f[i] <= '9' {
				n = n*10 + int(f[i]-'0')
				if n > math.MaxInt32 {
					panic(structError("total struct size too long"))
				}
				i++
			}
			if i >= len(f) {
				panic(structError("repeat count given without format specifier"))
			}
			count, hasCount = n, true
			c = f[i]
		}
		i++
		var size, align int
		var ok bool
		if sf.native {
			size, align, ok = nativeSize(c)
		} else {
			size, ok = stdSize(c)
			align = 1
		}
		if !ok {
			panic(structError("bad char in struct format"))
		}
		_ = hasCount
		if align > 1 {
			off = (off + align - 1) / align * align
		}
		it := structItem{code: c, count: count, size: size, offset: off}
		switch c {
		case 's', 'p':
			off += count
			sf.nvals++
		case 'x':
			off += count
		default:
			off += size * count
			sf.nvals += count
		}
		sf.items = append(sf.items, it)
	}
	sf.size = off
	structFmtCache[f] = sf
	return sf
}

func (sf *structFmt) order() binary.ByteOrder {
	if sf.little {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func structInt(v any) *big.Int {
	v = Unbox(v)
	if b, ok := bigOf(v); ok {
		if _, isBool := v.(bool); !isBool || true {
			return b
		}
	}
	if DynHasAttr(v, "__index__") {
		r := DynCall(DynGetAttr(v, "__index__"))
		if b, ok := bigOf(Unbox(r)); ok {
			return b
		}
	}
	panic(structError("required argument is not an integer"))
}

func rangeFor(code byte, native bool, size int) (lo, hi *big.Int, signed bool) {
	bits := uint(size * 8)
	signed = code == 'b' || code == 'h' || code == 'i' || code == 'l' || code == 'q' || code == 'n'
	if signed {
		hi = new(big.Int).Sub(new(big.Int).Lsh(bigOne, bits-1), bigOne)
		lo = new(big.Int).Neg(new(big.Int).Lsh(bigOne, bits-1))
	} else {
		lo = new(big.Int)
		hi = new(big.Int).Sub(new(big.Int).Lsh(bigOne, bits), bigOne)
	}
	return
}

func packInt(dst []byte, sf *structFmt, code byte, size int, v any) {
	n := structInt(v)
	lo, hi, _ := rangeFor(code, sf.native, size)
	if n.Cmp(lo) < 0 || n.Cmp(hi) > 0 {
		panic(structError("'" + string(code) + "' format requires " + lo.String() + " <= number <= " + hi.String()))
	}
	var u uint64
	if n.Sign() < 0 {
		u = uint64(n.Int64())
	} else {
		u = n.Uint64()
	}
	switch size {
	case 1:
		dst[0] = byte(u)
	case 2:
		sf.order().PutUint16(dst, uint16(u))
	case 4:
		sf.order().PutUint32(dst, uint32(u))
	case 8:
		sf.order().PutUint64(dst, u)
	}
}

func unpackInt(src []byte, sf *structFmt, code byte, size int) any {
	var u uint64
	switch size {
	case 1:
		u = uint64(src[0])
	case 2:
		u = uint64(sf.order().Uint16(src))
	case 4:
		u = uint64(sf.order().Uint32(src))
	case 8:
		u = sf.order().Uint64(src)
	}
	signed := code == 'b' || code == 'h' || code == 'i' || code == 'l' || code == 'q' || code == 'n'
	if signed {
		shift := uint(64 - size*8)
		return int64(u<<shift) >> shift
	}
	if u > math.MaxInt64 {
		return NormBig(new(big.Int).SetUint64(u))
	}
	return int64(u)
}

func float16bits(f float64) uint16 {
	b := math.Float32bits(float32(f))
	sign := uint16(b>>16) & 0x8000
	exp := int((b>>23)&0xff) - 127 + 15
	mant := b & 0x7fffff
	if math.IsNaN(f) {
		return sign | 0x7e00
	}
	if math.IsInf(f, 0) {
		return sign | 0x7c00
	}
	// round the double directly to half precision
	a := math.Abs(f)
	if a >= 65520 {
		panic(newError("OverflowError", "float too large to pack with e format"))
	}
	if a < 6.103515625e-05 { // subnormal half
		q := a / 5.960464477539063e-08
		r := math.RoundToEven(q)
		return sign | uint16(r)
	}
	frac, e := math.Frexp(a) // a = frac * 2**e, frac in [0.5,1)
	m := math.RoundToEven((frac*2 - 1) * 1024)
	ex := e - 1 + 15
	if m == 1024 {
		m = 0
		ex++
	}
	_ = exp
	_ = mant
	return sign | uint16(ex)<<10 | uint16(m)
}

func float16from(h uint16) float64 {
	sign := 1.0
	if h&0x8000 != 0 {
		sign = -1
	}
	exp := int(h>>10) & 0x1f
	mant := float64(h & 0x3ff)
	switch exp {
	case 0:
		return sign * mant * 5.960464477539063e-08
	case 31:
		if mant == 0 {
			return sign * math.Inf(1)
		}
		return math.NaN()
	}
	return sign * math.Ldexp(1+mant/1024, exp-15)
}

func structFloat(v any) float64 {
	v = Unbox(v)
	switch x := v.(type) {
	case float64:
		return x
	case bool, int64:
		return mathFloat(x)
	}
	if b, ok := bigOf(v); ok {
		return bigToFloat(b)
	}
	if DynHasAttr(v, "__float__") {
		return mathFloat(v)
	}
	panic(structError("required argument is not a float"))
}

func (sf *structFmt) pack(args []any) []byte {
	if len(args) != sf.nvals {
		panic(structError("pack expected " + strconv.Itoa(sf.nvals) + " items for packing (got " + strconv.Itoa(len(args)) + ")"))
	}
	buf := make([]byte, sf.size)
	ai := 0
	for _, it := range sf.items {
		off := it.offset
		switch it.code {
		case 'x':
			continue
		case 's', 'p':
			b, ok := bytesOf(args[ai])
			if !ok {
				panic(structError("argument for '" + string(it.code) + "' must be a bytes object"))
			}
			ai++
			if it.code == 's' {
				copy(buf[off:off+it.count], b)
			} else if it.count > 0 {
				n := len(b)
				if n > it.count-1 {
					n = it.count - 1
				}
				if n > 255 {
					n = 255
				}
				buf[off] = byte(n)
				copy(buf[off+1:off+it.count], b[:n])
			}
			continue
		}
		for k := 0; k < it.count; k++ {
			dst := buf[off+k*it.size : off+(k+1)*it.size]
			v := args[ai]
			ai++
			switch it.code {
			case 'c':
				b, ok := bytesOf(v)
				if _, isBytes := Unbox(v).(Bytes); !ok || !isBytes && !isByteArray(v) || len(b) != 1 {
					panic(structError("char format requires a bytes object of length 1"))
				}
				dst[0] = b[0]
			case '?':
				if DynBool(v) {
					dst[0] = 1
				}
			case 'f':
				f := structFloat(v)
				sf.order().PutUint32(dst, math.Float32bits(float32(f)))
				if math.IsInf(float64(float32(f)), 0) && !math.IsInf(f, 0) {
					panic(newError("OverflowError", "float too large to pack with f format"))
				}
			case 'd':
				sf.order().PutUint64(dst, math.Float64bits(structFloat(v)))
			case 'e':
				sf.order().PutUint16(dst, float16bits(structFloat(v)))
			case 'P':
				packInt(dst, sf, 'Q', 8, v)
			default:
				packInt(dst, sf, it.code, it.size, v)
			}
		}
	}
	return buf
}

func isByteArray(v any) bool { _, ok := Unbox(v).(*ByteArray); return ok }

func (sf *structFmt) unpack(data string) VarTuple[any] {
	out := make([]any, 0, sf.nvals)
	buf := []byte(data)
	for _, it := range sf.items {
		off := it.offset
		switch it.code {
		case 'x':
			continue
		case 's':
			out = append(out, Bytes(buf[off:off+it.count]))
			continue
		case 'p':
			if it.count == 0 {
				out = append(out, Bytes(""))
				continue
			}
			n := int(buf[off])
			if n > it.count-1 {
				n = it.count - 1
			}
			out = append(out, Bytes(buf[off+1:off+1+n]))
			continue
		}
		for k := 0; k < it.count; k++ {
			src := buf[off+k*it.size : off+(k+1)*it.size]
			switch it.code {
			case 'c':
				out = append(out, Bytes(src[:1]))
			case '?':
				out = append(out, src[0] != 0)
			case 'f':
				out = append(out, float64(math.Float32frombits(sf.order().Uint32(src))))
			case 'd':
				out = append(out, math.Float64frombits(sf.order().Uint64(src)))
			case 'e':
				out = append(out, float16from(sf.order().Uint16(src)))
			case 'P':
				out = append(out, unpackInt(src, sf, 'Q', 8))
			default:
				out = append(out, unpackInt(src, sf, it.code, it.size))
			}
		}
	}
	return NewVarTuple(out...)
}

func init() {
	RegisterNative("struct.pack", func(args []any, kw []KW) any {
		if len(args) < 1 {
			panic(newError("TypeError", "pack expected at least 1 argument, got 0"))
		}
		sf := parseStructFmt(structFmtArg(args[0]))
		return Bytes(sf.pack(args[1:]))
	})
	RegisterNative("struct.calcsize", func(args []any, kw []KW) any {
		nativeArgs("calcsize", args, 1, 1)
		return int64(parseStructFmt(structFmtArg(args[0])).size)
	})
	RegisterNative("struct.unpack", func(args []any, kw []KW) any {
		nativeArgs("unpack", args, 2, 2)
		sf := parseStructFmt(structFmtArg(args[0]))
		data := structBuffer(args[1])
		if len(data) != sf.size {
			panic(structError("unpack requires a buffer of " + strconv.Itoa(sf.size) + " bytes"))
		}
		return sf.unpack(data)
	})
	RegisterNative("struct.unpack_from", func(args []any, kw []KW) any {
		nativeArgs("unpack_from", args, 2, 3)
		sf := parseStructFmt(structFmtArg(args[0]))
		data := structBuffer(args[1])
		off := 0
		if len(args) > 2 {
			off = int(DynInt(args[2]))
		}
		for _, k := range kw {
			if k.Name == "offset" {
				off = int(DynInt(k.Val))
			}
		}
		if off < 0 {
			if off+len(data) < 0 {
				panic(structError("offset " + strconv.Itoa(off) + " out of range for " + strconv.Itoa(len(data)) + "-byte buffer"))
			}
			off += len(data)
		}
		if off > len(data) || len(data)-off < sf.size {
			if off > len(data) {
				panic(structError("offset " + strconv.Itoa(off) + " out of range for " + strconv.Itoa(len(data)) + "-byte buffer"))
			}
			panic(structError("unpack_from requires a buffer of at least " + strconv.Itoa(sf.size+off) + " bytes for unpacking " + strconv.Itoa(sf.size) + " bytes at offset " + strconv.Itoa(off) + " (actual buffer size is " + strconv.Itoa(len(data)) + ")"))
		}
		return sf.unpack(data[off : off+sf.size])
	})
	RegisterNative("struct.pack_into", func(args []any, kw []KW) any {
		nativeArgs("pack_into", args, 3, 1<<20)
		sf := parseStructFmt(structFmtArg(args[0]))
		ba, ok := Unbox(args[1]).(*ByteArray)
		if !ok {
			panic(newError("TypeError", "argument must be read-write bytes-like object, not "+TypeName(Unbox(args[1]))))
		}
		off := int(DynInt(args[2]))
		if off < 0 {
			if off+len(ba.b) < 0 {
				panic(structError("offset " + strconv.Itoa(off) + " out of range for " + strconv.Itoa(len(ba.b)) + "-byte buffer"))
			}
			off += len(ba.b)
		}
		packed := sf.pack(args[3:])
		if off > len(ba.b) || len(ba.b)-off < sf.size {
			panic(structError("pack_into requires a buffer of at least " + strconv.Itoa(sf.size+off) + " bytes for packing " + strconv.Itoa(sf.size) + " bytes at offset " + strconv.Itoa(off) + " (actual buffer size is " + strconv.Itoa(len(ba.b)) + ")"))
		}
		copy(ba.b[off:], packed)
		return nil
	})
	RegisterNative("struct.nvals", func(args []any, kw []KW) any {
		return int64(parseStructFmt(structFmtArg(args[0])).nvals)
	})
}

func structFmtArg(v any) string {
	v = Unbox(v)
	switch x := v.(type) {
	case string:
		return x
	case Bytes:
		return string(x)
	}
	panic(newError("TypeError", "Struct() argument 1 must be a str or bytes object, not "+TypeName(v)))
}

func structBuffer(v any) string {
	s, ok := bytesOf(v)
	if !ok {
		panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(Unbox(v))+"'"))
	}
	return s
}

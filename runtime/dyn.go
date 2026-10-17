package runtime

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// Dynamic operations. A value of static type object is a Go any holding an
// int64, float64, string, bool, a container, a function or an instance of a
// class of the program, with None as nil. The operations here follow Python's
// rules for those values at run time.

// IsNone reports whether v is None: the nil of an object, the NoneT of a typed
// None, or an Optional that holds nothing.
func IsNone(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case NoneT:
		return true
	case interface{ IsSomeAny() bool }:
		return !x.IsSomeAny()
	}
	return false
}

// Unbox turns an Optional held in an object into its value or nil.
func Unbox(v any) any {
	if o, ok := v.(interface{ Boxed() any }); ok {
		return o.Boxed()
	}
	if _, ok := v.(NoneT); ok {
		return nil
	}
	if e, ok := v.(*PyError); ok && e == nil {
		return nil
	}
	return v
}

type numKind int

const (
	notNum numKind = iota
	isInt
	isFloat
)

// numeric reads an int, a bool or a float.
func numeric(v any) (numKind, int64, float64) {
	switch x := v.(type) {
	case int64:
		return isInt, x, 0
	case int:
		return isInt, int64(x), 0
	case bool:
		return isInt, b2i(x), 0
	case float64:
		return isFloat, 0, x
	}
	return notNum, 0, 0
}

func asFloat(k numKind, i int64, f float64) float64 {
	if k == isInt {
		return float64(i)
	}
	return f
}

// DynBinary is a op b for the binary operators, named as in the AST: Add, Sub,
// Mult, Div, FloorDiv, Mod, Pow, MatMult, BitAnd, BitOr, BitXor, LShift,
// RShift.
func DynBinary(op string, a, b any) any {
	a, b = Unbox(a), Unbox(b)
	// Instances of classes try their own methods, then the other operand's
	// reflected ones.
	if x, ok := a.(binaryOperand); ok {
		if r, ok := x.PyBinop(op, b, false); ok {
			return r
		}
	}
	if y, ok := b.(binaryOperand); ok && !sameClass(a, b) {
		if r, ok := y.PyBinop(op, a, true); ok {
			return r
		}
	}
	abool, aIsBool := a.(bool)
	bbool, bIsBool := b.(bool)
	if aIsBool && bIsBool {
		switch op {
		case "BitAnd":
			return abool && bbool
		case "BitOr":
			return abool || bbool
		case "BitXor":
			return abool != bbool
		}
	}
	if aIsBool {
		a = b2i(abool)
	}
	if bIsBool {
		b = b2i(bbool)
	}
	if isIntVal(a) && isIntVal(b) {
		if r, ok := intArith(op, a, b); ok {
			return r
		}
	}
	if isComplexVal(a) || isComplexVal(b) {
		x, okx := toComplex(a)
		y, oky := toComplex(b)
		if okx && oky {
			if r := complexBinary(op, x, y); r != nil {
				return r
			}
			panic(UnsupportedOperand(opSymbol(op), TypeName(a), TypeName(b)))
		}
	}
	// A big int next to a float is a float.
	if isBig(a) {
		if f, ok := b.(float64); ok {
			return DynBinary(op, bigToFloat(a.(*big.Int)), f)
		}
	}
	if isBig(b) {
		if f, ok := a.(float64); ok {
			return DynBinary(op, f, bigToFloat(b.(*big.Int)))
		}
	}
	ka, ia, fa := numeric(a)
	kb, ib, fb := numeric(b)
	if ka != notNum && kb != notNum {
		return numericBinary(op, ka, ia, fa, kb, ib, fb)
	}
	switch x := a.(type) {
	case string:
		switch y := b.(type) {
		case string:
			if op == "Add" {
				return x + y
			}
		default:
			if n, ok := intOf(b); ok && op == "Mult" {
				return StrRepeat(x, n)
			}
		}
		if op == "Mod" {
			return PercentFormat(x, b)
		}
	case seqOperand:
		if r, ok := x.DynBinary(op, b, false); ok {
			return r
		}
	}
	if y, ok := b.(seqOperand); ok {
		if r, ok := y.DynBinary(op, a, true); ok {
			return r
		}
	}
	if s, ok := b.(string); ok && op == "Mult" {
		if n, ok := intOf(a); ok {
			return StrRepeat(s, n)
		}
	}
	if r, ok := seqFallback(op, a, b); ok {
		return r
	}
	if op == "Mult" {
		for _, pair := range [][2]any{{a, b}, {b, a}} {
			_, isSeq := pair[0].(Seq)
			_, isStr := pair[0].(string)
			bs, isBytes := bytesOf(pair[0])
			if isBig(pair[1]) && (isSeq || isStr || isBytes) {
				empty := isBytes && bs == ""
				if str, ok := pair[0].(string); ok {
					empty = str == ""
				}
				if sq, ok := pair[0].(Seq); ok {
					empty = sq.Len() == 0
				}
				if empty {
					return DynBinary(op, pair[0], int64(0))
				}
				panic(overflowError("cannot fit 'int' into an index-sized integer"))
			}
			if k, _, _ := numeric(pair[1]); (isSeq || isStr || isBytes) && k != isInt && !isBig(pair[1]) {
				panic(newError("TypeError", "can't multiply sequence by non-int of type '"+TypeName(pair[1])+"'"))
			}
		}
	}
	if op == "Add" {
		_, aSeq := a.(Seq)
		_, aStr := a.(string)
		if aStr || aSeq {
			panic(newError("TypeError", "can only concatenate "+TypeName(a)+" (not \""+TypeName(b)+"\") to "+TypeName(a)))
		}
	}
	panic(UnsupportedOperand(opSymbol(op), TypeName(a), TypeName(b)))
}

// seqFallback concatenates and repeats sequences whose item types differ.
func seqFallback(op string, a, b any) (any, bool) {
	sa, oka := a.(Seq)
	sb, okb := b.(Seq)
	items := func(s Seq) []any {
		out := make([]any, s.Len())
		for i := range out {
			out[i] = s.At(int64(i))
		}
		return out
	}
	build := func(kind string, xs []any) (any, bool) {
		switch kind {
		case "tuple":
			return NewVarTuple[any](xs...), true
		case "list":
			return &List[any]{items: xs}, true
		}
		return nil, false
	}
	switch {
	case op == "Add" && oka && okb && sa.Kind() == sb.Kind():
		return build(sa.Kind(), append(items(sa), items(sb)...))
	case op == "Mult" && oka:
		if n, ok := intOf(b); ok {
			var out []any
			for i := int64(0); i < n; i++ {
				out = append(out, items(sa)...)
			}
			return build(sa.Kind(), out)
		}
	case op == "Mult" && okb:
		if n, ok := intOf(a); ok {
			var out []any
			for i := int64(0); i < n; i++ {
				out = append(out, items(sb)...)
			}
			return build(sb.Kind(), out)
		}
	}
	return nil, false
}

// binaryOperand is implemented by instances whose class defines operator
// methods. The result is false when no method took the operands.
type binaryOperand interface {
	PyBinop(op string, other any, reflected bool) (any, bool)
}

// seqOperand is implemented by the containers that support operators.
type seqOperand interface {
	DynBinary(op string, other any, reflected bool) (any, bool)
}

func sameClass(a, b any) bool {
	na, oka := a.(Named)
	nb, okb := b.(Named)
	return oka && okb && na.PyClassName() == nb.PyClassName()
}

func intOf(v any) (int64, bool) {
	k, i, _ := numeric(v)
	return i, k == isInt
}

var opSymbols = map[string]string{
	"Add": "+", "Sub": "-", "Mult": "*", "Div": "/", "FloorDiv": "//", "Mod": "%", "Pow": "**", "MatMult": "@",
	"BitAnd": "&", "BitOr": "|", "BitXor": "^", "LShift": "<<", "RShift": ">>",
}

func opSymbol(op string) string { return opSymbols[op] }

func numericBinary(op string, ka numKind, ia int64, fa float64, kb numKind, ib int64, fb float64) any {
	if ka == isInt && kb == isInt {
		switch op {
		case "Add":
			return addInt(ia, ib)
		case "Sub":
			return subInt(ia, ib)
		case "Mult":
			return mulInt(ia, ib)
		case "Div":
			return DivInt(ia, ib)
		case "FloorDiv":
			return FloorDiv(ia, ib)
		case "Mod":
			return Mod(ia, ib)
		case "Pow":
			if ib >= 0 {
				return PowInt(ia, ib)
			}
			return PowFloat(float64(ia), float64(ib))
		case "BitAnd":
			return ia & ib
		case "BitOr":
			return ia | ib
		case "BitXor":
			return ia ^ ib
		case "LShift":
			return LShift(ia, ib)
		case "RShift":
			return RShift(ia, ib)
		}
		panic(UnsupportedOperand(opSymbol(op), "int", "int"))
	}
	x, y := asFloat(ka, ia, fa), asFloat(kb, ib, fb)
	switch op {
	case "Add":
		return x + y
	case "Sub":
		return x - y
	case "Mult":
		return x * y
	case "Div":
		return Div(x, y)
	case "FloorDiv":
		return FloorDivFloat(x, y)
	case "Mod":
		return ModFloat(x, y)
	case "Pow":
		if x < 0 && !math.IsInf(x, 0) && !math.IsInf(y, 0) && y != math.Trunc(y) {
			return complexPow(complex(x, 0), complex(y, 0))
		}
		return PowFloat(x, y)
	}
	panic(UnsupportedOperand(opSymbol(op), kindName(ka), kindName(kb)))
}

// The int64 arithmetic does not overflow silently: Python's ints are
// unbounded, so a result that does not fit is an OverflowError until phase 10.
func addInt(a, b int64) int64 {
	c := a + b
	if (c > a) != (b > 0) {
		panic(overflowError("int too large for int64 (big integers arrive in phase 10)"))
	}
	return c
}

func subInt(a, b int64) int64 {
	c := a - b
	if (c < a) != (b > 0) {
		panic(overflowError("int too large for int64 (big integers arrive in phase 10)"))
	}
	return c
}

func mulInt(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	c := a * b
	if c/b != a || (a == math.MinInt64 && b == -1) || (b == math.MinInt64 && a == -1) {
		panic(overflowError("int too large for int64 (big integers arrive in phase 10)"))
	}
	return c
}

// DynUnary is -x, +x, ~x and abs(x): op is "USub", "UAdd", "Invert" or "Abs".
func DynUnary(op string, a any) any {
	a = Unbox(a)
	if x, ok := a.(unaryOperand); ok {
		if r, ok := x.PyUnop(op); ok {
			return r
		}
	}
	if i, ok := a.(int64); ok && i != math.MinInt64 {
		switch op {
		case "USub":
			return -i
		case "UAdd":
			return i
		case "Invert":
			return ^i
		case "Abs":
			return AbsInt(i)
		}
	}
	if isIntVal(a) {
		if r, ok := intUnary(op, a); ok {
			return r
		}
	}
	if z, ok := a.(complex128); ok {
		switch op {
		case "USub":
			return -z
		case "UAdd":
			return z
		case "Abs":
			return math.Hypot(real(z), imag(z))
		}
	}
	switch k, i, f := numeric(a); k {
	case isInt:
		switch op {
		case "USub":
			return -i
		case "UAdd":
			return i
		case "Invert":
			return ^i
		case "Abs":
			return AbsInt(i)
		}
	case isFloat:
		switch op {
		case "USub":
			return -f
		case "UAdd":
			return f
		case "Abs":
			return math.Abs(f)
		}
	}
	name := map[string]string{"USub": "-", "UAdd": "+", "Invert": "~"}[op]
	if op == "Abs" {
		panic(newError("TypeError", "bad operand type for abs(): '"+TypeName(a)+"'"))
	}
	panic(UnsupportedOperandUnary(name, TypeName(a)))
}

type unaryOperand interface {
	PyUnop(op string) (any, bool)
}

// DynNot is "not x".
func DynNot(a any) bool { return !Truth(Unbox(a)) }

// DynCompare is a op b for the comparison operators named Eq, NotEq, Lt, LtE,
// Gt and GtE.
func DynCompare(op string, a, b any) bool {
	a, b = Unbox(a), Unbox(b)
	switch op {
	case "Eq":
		return Eq(a, b)
	case "NotEq":
		return !Eq(a, b)
	case "Lt":
		return Lt(a, b)
	case "LtE":
		return Le(a, b)
	case "Gt":
		return Gt(a, b)
	}
	return Ge(a, b)
}

// DynContains is "item in container".
func DynContains(container, item any) bool {
	container, item = Unbox(container), Unbox(item)
	switch c := container.(type) {
	case string:
		s, ok := item.(string)
		if !ok {
			panic(newError("TypeError", "'in <string>' requires string as left operand, not "+TypeName(item)))
		}
		return strings.Contains(c, s)
	case containsAny:
		return c.ContainsAny(item)
	case interface{ PyContains(item any) (bool, bool) }:
		if r, ok := c.PyContains(item); ok {
			return r
		}
	}
	if s, ok := container.(Seq); ok {
		for i := int64(0); i < s.Len(); i++ {
			if Eq(s.At(i), item) {
				return true
			}
		}
		return false
	}
	if it, ok := container.(interface{ PyIterAny() (Cursor[any], bool) }); ok {
		if cur, ok := it.PyIterAny(); ok {
			for cur.Next() {
				if Eq(cur.Value(), item) {
					return true
				}
			}
			return false
		}
	}
	// Anything iterable is searched.
	if it, ok := container.(interface{ IterAny() Cursor[any] }); ok {
		for cur := it.IterAny(); cur.Next(); {
			if Eq(cur.Value(), item) {
				return true
			}
		}
		return false
	}
	panic(newError("TypeError", "argument of type '"+TypeName(container)+"' is not iterable"))
}

type containsAny interface{ ContainsAny(item any) bool }

// DynLen is len(x).
func DynLen(v any) int64 {
	v = Unbox(v)
	switch x := v.(type) {
	case string:
		return StrLen(x)
	case interface{ Len() int64 }:
		return x.Len()
	case interface{ PyLen() (int64, bool) }:
		if n, ok := x.PyLen(); ok {
			return n
		}
	}
	panic(newError("TypeError", "object of type '"+TypeName(v)+"' has no len()"))
}

// DynInt is int(x) for a result that fits a machine word; DynIntAny is the
// full int(x).
func DynInt(v any) int64 {
	r := DynIntAny(v)
	if n, ok := r.(int64); ok {
		return n
	}
	panic(overflowError("Python int too large to convert to C long"))
}

// DynIntAny is int(x): ints of any size, floats, strings and objects with
// __int__.
func DynIntAny(v any) any {
	v = BaseValue(Unbox(v))
	switch x := v.(type) {
	case int64:
		return x
	case *big.Int:
		return x
	case bool:
		return b2i(x)
	case float64:
		switch {
		case math.IsNaN(x):
			panic(valueError("cannot convert float NaN to integer"))
		case math.IsInf(x, 0):
			panic(overflowError("cannot convert float infinity to integer"))
		}
		x = math.Trunc(x)
		if x >= 1<<63 || x < -(1<<63) {
			return bigFromFloat(x)
		}
		return int64(x)
	case string:
		t := asciiDigits(StrStrip(x))
		sign := ""
		body := t
		if body != "" && (body[0] == '+' || body[0] == '-') {
			sign, body = body[:1], body[1:]
		}
		if !digitPart(body) {
			panic(valueError("invalid literal for int() with base 10: " + quote(x)))
		}
		return IntFromText(sign+body, 10, x)
	case Bytes:
		return DynIntAny(string(x))
	case *ByteArray:
		return DynIntAny(string(x.b))
	case interface{ PyInt() (any, bool) }:
		if n, ok := x.PyInt(); ok {
			return n
		}
	}
	panic(newError("TypeError", "int() argument must be a string, a bytes-like object or a real number, not '"+TypeName(v)+"'"))
}

// DynFloat is float(x).
func DynFloat(v any) float64 {
	v = BaseValue(Unbox(v))
	switch x := v.(type) {
	case *big.Int:
		return bigToFloat(x)
	case int64:
		return float64(x)
	case bool:
		return float64(b2i(x))
	case float64:
		return x
	case string:
		return FloatFromStr(x)
	case Bytes:
		return FloatFromStr(string(x))
	case *ByteArray:
		return FloatFromStr(string(x.b))
	case interface{ PyFloat() (float64, bool) }:
		if f, ok := x.PyFloat(); ok {
			return f
		}
	}
	panic(newError("TypeError", "float() argument must be a string or a real number, not '"+TypeName(v)+"'"))
}

// DynRound is round(x) and round(x, n).
func DynRound(v any, digits any) any {
	v = Unbox(v)
	d, hasDigits := int64(0), digits != nil && !IsNone(digits)
	if hasDigits {
		dv := Unbox(digits)
		if _, isInt := intOf(dv); !isInt {
			panic(newError("TypeError", "'"+TypeName(dv)+"' object cannot be interpreted as an integer"))
		}
		d = DynInt(digits)
	}
	if b, ok := v.(*big.Int); ok {
		if !hasDigits || d >= 0 {
			return b
		}
		p := new(big.Int).Exp(big.NewInt(10), big.NewInt(-d), nil)
		q, r := bigFloorDivMod(b, p, "")
		twice := new(big.Int).Lsh(r, 1)
		switch c := twice.Cmp(p); {
		case c > 0, c == 0 && q.Bit(0) == 1:
			q.Add(q, bigOne)
		}
		return NormBig(q.Mul(q, p))
	}
	switch k, i, f := numeric(v); k {
	case isInt:
		if !hasDigits || d >= 0 {
			return i
		}
		p := math.Pow(10, float64(-d))
		return int64(math.RoundToEven(float64(i)/p) * p)
	case isFloat:
		if !hasDigits {
			return DynIntAny(math.RoundToEven(f))
		}
		p := math.Pow(10, float64(d))
		if math.IsInf(f, 0) || math.IsNaN(f) || math.IsInf(p, 0) {
			return f
		}
		return RoundDigits(f, d)
	}
	if _, named := v.(Named); named && DynHasAttr(v, "__round__") {
		if hasDigits {
			return DynCall(DynGetAttr(v, "__round__"), digits)
		}
		return DynCall(DynGetAttr(v, "__round__"))
	}
	panic(newError("TypeError", "type "+TypeName(v)+" doesn't define __round__ method"))
}

// RoundDigits is round(x, n) for a float: correctly rounded to n decimal
// digits, with ties going to even on the exact decimal value.
func RoundDigits(f float64, n int64) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) || f == 0 {
		return f
	}
	if n >= 0 {
		if n > 400 {
			return f
		}
		out, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'f', int(n), 64), 64)
		return out
	}
	// Round the exact value to a multiple of a power of ten, ties to even.
	if -n > 400 {
		return math.Copysign(0, f)
	}
	p := new(big.Int).Exp(big.NewInt(10), big.NewInt(-n), nil)
	r := new(big.Rat).SetFloat64(f)
	r.Quo(r, new(big.Rat).SetInt(p))
	num, den := r.Num(), r.Denom()
	q, rem := new(big.Int).DivMod(num, den, new(big.Int))
	twice := new(big.Int).Lsh(rem, 1)
	switch c := twice.Cmp(den); {
	case c > 0, c == 0 && q.Bit(0) == 1:
		q.Add(q, bigOne)
	}
	out, _ := new(big.Rat).SetInt(q.Mul(q, p)).Float64()
	if out == 0 {
		return math.Copysign(0, f)
	}
	return out
}

// DynStr is str(x); DynRepr is repr(x).
func DynStr(v any) string  { return Str(Unbox(v)) }
func DynRepr(v any) string { return Repr(Unbox(v)) }

// DynBool is bool(x).
func DynBool(v any) bool { return Truth(Unbox(v)) }

// PercentFormat is "fmt" % args for the common conversions: %s %r %d %i %f
// %x %o %c %%, with flags, width and precision.
func PercentFormat(format string, args any) string {
	var list []any
	switch a := args.(type) {
	case Seq:
		if a.Kind() == "tuple" {
			for i := int64(0); i < a.Len(); i++ {
				list = append(list, a.At(i))
			}
			break
		}
		list = []any{args}
	default:
		list = []any{args}
	}
	var b strings.Builder
	next := 0
	usedMapping := false
	take := func() any {
		if next >= len(list) {
			panic(newError("TypeError", "not enough arguments for format string"))
		}
		v := list[next]
		next++
		return v
	}
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' {
			b.WriteByte(c)
			continue
		}
		i++
		if i >= len(format) {
			panic(valueError("incomplete format"))
		}
		hasKey := false
		var mapKey string
		if format[i] == '(' {
			end := strings.IndexByte(format[i:], ')')
			if end < 0 {
				panic(valueError("incomplete format key"))
			}
			mapKey, hasKey = format[i+1:i+end], true
			if seq, isSeq := args.(Seq); isSeq && seq.Kind() != "dict" {
				panic(newError("TypeError", "format requires a mapping"))
			}
			if !isSeqOrMapping(args) {
				panic(newError("TypeError", "format requires a mapping"))
			}
			usedMapping = true
			i += end + 1
			if i >= len(format) {
				panic(valueError("incomplete format"))
			}
		}
		take := func() any {
			if hasKey {
				return DynGetItem(args, mapKey)
			}
			if usedMapping {
				panic(newError("TypeError", "not enough arguments for format string"))
			}
			return take()
		}
		start := i
		for i < len(format) && strings.IndexByte("-+ #0", format[i]) >= 0 {
			i++
		}
		spec := format[start:i]
		for i < len(format) && (format[i] >= '0' && format[i] <= '9' || format[i] == '.' || format[i] == '*') {
			if format[i] == '*' {
				spec += strconv.FormatInt(DynInt(take()), 10)
			} else {
				spec += string(format[i])
			}
			i++
		}
		if i >= len(format) {
			panic(valueError("incomplete format"))
		}
		verb := format[i]
		goSpec := "%" + spec
		switch verb {
		case '%':
			b.WriteByte('%')
		case 's':
			fmt.Fprintf(&b, goSpec+"s", Str(Unbox(take())))
		case 'r':
			fmt.Fprintf(&b, goSpec+"s", Repr(Unbox(take())))
		case 'd', 'i', 'u':
			v := Unbox(take())
			if f, ok := v.(float64); ok {
				v = IntFromFloat(f)
			}
			if x, ok := v.(interface{ PyIndex() (any, bool) }); ok {
				if r, ok := x.PyIndex(); ok {
					v = Unbox(r)
				}
			} else if x, ok := v.(interface{ PyInt() (any, bool) }); ok {
				if r, ok := x.PyInt(); ok {
					v = Unbox(r)
				}
			}
			if bi, isBig := v.(*big.Int); isBig {
				b.WriteString(FormatBig(bi, strings.Replace(strings.TrimLeft(goSpec, "%"), "-", "<", 1)+"d"))
				break
			}
			n, ok := intOf(v)
			if !ok {
				panic(newError("TypeError", "%d format: a real number is required, not "+TypeName(v)))
			}
			fmt.Fprintf(&b, goSpec+"d", n)
		case 'f', 'F', 'e', 'E', 'g', 'G':
			v := Unbox(take())
			if bi, isBig := v.(*big.Int); isBig {
				v = bigToFloat(bi)
			}
			k, i2, f := numeric(v)
			if k == notNum {
				panic(newError("TypeError", "must be real number, not "+TypeName(v)))
			}
			if ff := asFloat(k, i2, f); math.IsInf(ff, 0) || math.IsNaN(ff) {
				b.WriteString(padNonFinite(goSpec[1:], ff, rune(verb)))
				break
			}
			spec := goSpec
			if (verb == 'g' || verb == 'G') && !strings.Contains(spec, ".") {
				spec += ".6"
			}
			fmt.Fprintf(&b, spec+string(verb), asFloat(k, i2, f))
		case 'x', 'X', 'o':
			v := Unbox(take())
			if bi, isBig := v.(*big.Int); isBig {
				b.WriteString(FormatBig(bi, strings.Replace(strings.TrimLeft(goSpec, "%"), "-", "<", 1)+string(verb)))
				break
			}
			n, ok := intOf(v)
			if !ok {
				panic(newError("TypeError", "%"+string(verb)+" format: an integer is required, not "+TypeName(v)))
			}
			if verb == 'o' && strings.Contains(goSpec, "#") {
				sign := ""
				if n < 0 {
					sign = "-"
				}
				plain := strings.Replace(goSpec, "#", "", 1)
				b.WriteString(fmt.Sprintf(plain[:len(plain)]+"s", sign+"0o"+strconv.FormatUint(absUint(n), 8)))
				break
			}
			fmt.Fprintf(&b, goSpec+string(verb), n)
		case 'a':
			fmt.Fprintf(&b, goSpec+"s", Ascii(Unbox(take())))
		case 'c':
			v := Unbox(take())
			if n, ok := intOf(v); ok && !isBig(v) {
				if n < 0 || n > 0x10ffff {
					panic(overflowError("%c arg not in range(0x110000)"))
				}
				b.WriteRune(rune(n))
			} else if cs, isStr := v.(string); isStr && utf8.RuneCountInString(cs) == 1 {
				b.WriteString(cs)
			} else {
				panic(newError("TypeError", "%c requires int or char"))
			}
		default:
			panic(valueError(fmt.Sprintf("unsupported format character '%c' (0x%x) at index %d", verb, verb, utf8.RuneCountInString(format[:i]))))
		}
	}
	_, isSeq := args.(Seq)
	_, mapping := args.(interface{ GetItemAny(any) any })
	if next < len(list) && !usedMapping && !(mapping && !isSeq) {
		panic(newError("TypeError", "not all arguments converted during string formatting"))
	}
	return b.String()
}

// DynInplace is a op= b: lists change in place for += and *=, everything
// else is the plain operator.
func DynInplace(op string, a, b any) any {
	a, b = Unbox(a), Unbox(b)
	if x, ok := a.(interface {
		PyInplace(op string, o any) (any, bool)
	}); ok {
		if r, ok := x.PyInplace(op, b); ok {
			return r
		}
	}
	if l, ok := a.(interface{ InplaceBinary(op string, b any) bool }); ok && l.InplaceBinary(op, b) {
		return a
	}
	return DynBinary(op, a, b)
}

// normKey makes the keys 1, 1.0 and True of a dict with object keys one key,
// and refuses the ones that cannot be hashed.
func normKey[K comparable](k K) K {
	if p, ok := any(&k).(*any); ok {
		*p = dynKey(*p)
	}
	return k
}

func dynKey(v any) any {
	if b, ok := baseValueOf(Unbox(v)); ok && immutableBase(b) {
		// A str or int subclass is the same key as its value.
		v = b
	}
	switch x := Unbox(v).(type) {
	case bool:
		return b2i(x)
	case complex128:
		if imag(x) == 0 {
			return dynKey(real(x))
		}
		return x
	case float64:
		if x == math.Trunc(x) && !math.IsInf(x, 0) {
			if x >= -(1<<63) && x < 1<<63 {
				return int64(x)
			}
			return bigKey(bigFromFloat(x).(*big.Int).String())
		}
		return x
	case int64, string, nil:
		return x
	case *big.Int:
		return bigKey(x.String())
	case *FrozenSet:
		var parts []string
		for c := x.s.Iter(); c.Next(); {
			var b strings.Builder
			canonKey(&b, c.Value())
			parts = append(parts, b.String())
		}
		sort.Strings(parts)
		return tupleKey("F{" + strings.Join(parts, "") + "}")
	case Seq:
		if x.Kind() == "tuple" {
			var b strings.Builder
			canonKey(&b, x)
			return tupleKey(b.String())
		}
		Hash(x)
		return x
	case userHasher:
		return internUserKey(x)
	default:
		Hash(x) // raises TypeError for an unhashable value
		return x
	}
}

// userHasher is implemented by the instances of a class that defines
// __hash__.
type userHasher interface {
	Hasher
	PyUserHash()
}

// userKey stands for an object with its own __hash__ among the keys of a
// dict or the elements of a set: equal objects (by __eq__) get one key.
type userKey struct {
	h int64
	n int
}

var (
	userKeysMu sync.Mutex
	userKeys   = map[int64][]any{}
)

func internUserKey(x userHasher) userKey {
	h := Hash(x)
	userKeysMu.Lock()
	reps := userKeys[h]
	userKeysMu.Unlock()
	for i, rep := range reps {
		if Eq(rep, x) {
			return userKey{h, i}
		}
	}
	userKeysMu.Lock()
	defer userKeysMu.Unlock()
	userKeys[h] = append(userKeys[h], x)
	return userKey{h, len(userKeys[h]) - 1}
}

// bigKey stands for an int beyond 64 bits among the keys of a dict.
type bigKey string

// tupleKey stands for a tuple among the keys of a dict whose keys are objects:
// the tuple itself cannot be a Go map key.
type tupleKey string

// canonKey writes the text two equal keys have in common.
func canonKey(b *strings.Builder, v any) {
	if s, ok := Unbox(v).(Seq); ok && s.Kind() == "tuple" {
		b.WriteByte('(')
		for i := int64(0); i < s.Len(); i++ {
			canonKey(b, s.At(i))
		}
		b.WriteByte(')')
		return
	}
	switch x := dynKey(v).(type) {
	case nil:
		b.WriteString("n;")
	case int64:
		fmt.Fprintf(b, "i%d;", x)
	case float64:
		fmt.Fprintf(b, "f%v;", x)
	case string:
		fmt.Fprintf(b, "s%d:%s;", len(x), x)
	case tupleKey:
		b.WriteString(string(x))
	case bigKey:
		b.WriteString("b" + string(x) + ";")
	case userKey:
		fmt.Fprintf(b, "u%d.%d;", x.h, x.n)
	default:
		fmt.Fprintf(b, "p%p;", x)
	}
}

// keyEq compares two keys of a set: Go equality, and for object keys the
// Python one, under which 1, 1.0 and True are the same.
func keyEq[T comparable](a, b T) bool {
	if _, ok := any(&a).(*any); ok {
		return Eq(any(a), any(b))
	}
	return a == b
}

func isFloatValue(v any) bool { _, ok := v.(float64); return ok }

func kindName(k numKind) string {
	if k == isInt {
		return "int"
	}
	return "float"
}

// padNonFinite is %f, %e or %g of an infinity or a NaN: Python writes them as
// inf and nan (INF and NAN for the upper-case verbs), where Go writes +Inf.
func padNonFinite(spec string, f float64, verb rune) string {
	text := "inf"
	if math.IsNaN(f) {
		text = "nan"
	}
	if verb == 'F' || verb == 'E' || verb == 'G' {
		text = strings.ToUpper(text)
	}
	switch {
	case math.IsNaN(f):
	case f < 0:
		text = "-" + text
	case strings.Contains(spec, "+"):
		text = "+" + text
	case strings.Contains(spec, " "):
		text = " " + text
	}
	i := 0
	for i < len(spec) && strings.IndexByte("-+ #0", spec[i]) >= 0 {
		i++
	}
	j := i
	for j < len(spec) && spec[j] >= '0' && spec[j] <= '9' {
		j++
	}
	width, _ := strconv.Atoi(spec[i:j])
	if n := width - len(text); n > 0 {
		if strings.Contains(spec[:i], "-") {
			return text + strings.Repeat(" ", n)
		}
		return strings.Repeat(" ", n) + text
	}
	return text
}

// isSeqOrMapping is false for the plain scalars, which are not mappings.
func isSeqOrMapping(v any) bool {
	switch v.(type) {
	case nil, int64, float64, bool, string, *big.Int, complex128:
		return false
	}
	return true
}

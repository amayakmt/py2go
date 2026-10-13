package runtime

import (
	"cmp"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// BoolToInt is int(True) and int(False).
func BoolToInt(b bool) int64 { return b2i(b) }

// IntFromFloat is int(x): truncation toward zero. Python ints are
// unbounded, so a result outside int64 is an OverflowError here.
func IntFromFloat(f float64) int64 {
	switch {
	case math.IsNaN(f):
		panic(valueError("cannot convert float NaN to integer"))
	case math.IsInf(f, 0):
		panic(overflowError("cannot convert float infinity to integer"))
	}
	f = math.Trunc(f)
	if f >= 1<<63 || f < -(1<<63) {
		panic(overflowError("int too large to convert to int64"))
	}
	return int64(f)
}

// RoundFloat is round(x) with one argument: ties go to the even integer.
func RoundFloat(f float64) int64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return IntFromFloat(f)
	}
	return IntFromFloat(math.RoundToEven(f))
}

// digitPart reports whether s is digits with single underscores between
// digits, the shape Python allows in numeric literals and conversions.
func digitPart(s string) bool {
	if s == "" || s[0] == '_' || s[len(s)-1] == '_' || strings.Contains(s, "__") {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != '_' && (s[i] < '0' || s[i] > '9') {
			return false
		}
	}
	return true
}

// IntFromStr is int(s) in base 10.
func IntFromStr(s string) int64 {
	t := asciiDigits(StrStrip(s))
	neg := false
	if t != "" && (t[0] == '+' || t[0] == '-') {
		neg = t[0] == '-'
		t = t[1:]
	}
	if !digitPart(t) {
		panic(valueError("invalid literal for int() with base 10: " + quote(s)))
	}
	n, err := strconv.ParseUint(strings.ReplaceAll(t, "_", ""), 10, 64)
	if err != nil || (!neg && n > math.MaxInt64) || (neg && n > 1<<63) {
		panic(overflowError("int too large to convert to int64"))
	}
	if neg {
		return -int64(n)
	}
	return int64(n)
}

// FloatFromStr is float(s): optional sign, then inf, infinity, nan, or a
// decimal number with optional fraction and exponent. Underscores may
// separate digits. Go's parser alone is too permissive (hex floats).
func FloatFromStr(s string) float64 {
	t := asciiDigits(StrStrip(s))
	body := t
	if body != "" && (body[0] == '+' || body[0] == '-') {
		body = body[1:]
	}
	switch strings.ToLower(body) {
	case "inf", "infinity":
		f, _ := strconv.ParseFloat(t, 64)
		return f
	case "nan":
		return math.NaN()
	}

	mant, exp, hasExp := strings.Cut(strings.ToLower(body), "e")
	intPart, frac, hasDot := strings.Cut(mant, ".")
	ok := (intPart != "" || frac != "") &&
		(intPart == "" || digitPart(intPart)) &&
		(!hasDot || frac == "" || digitPart(frac))
	if hasExp {
		if exp != "" && (exp[0] == '+' || exp[0] == '-') {
			exp = exp[1:]
		}
		ok = ok && digitPart(exp)
	}
	if !ok {
		panic(valueError("could not convert string to float: " + quote(s)))
	}
	f, _ := strconv.ParseFloat(strings.ReplaceAll(t, "_", ""), 64) // overflow gives ±Inf, as in Python
	return f
}

func AbsInt(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func AbsFloat(x float64) float64 { return math.Abs(x) }

// Min and Max follow Python's rule of keeping the first of equal items,
// which matters for floats: a NaN never replaces the current best.
func Min[T cmp.Ordered](first T, rest ...T) T {
	best := first
	for _, x := range rest {
		if x < best {
			best = x
		}
	}
	return best
}

func Max[T cmp.Ordered](first T, rest ...T) T {
	best := first
	for _, x := range rest {
		if x > best {
			best = x
		}
	}
	return best
}

func MinList[T cmp.Ordered](l *List[T]) T {
	if len(l.items) == 0 {
		panic(valueError("min() iterable argument is empty"))
	}
	return Min(l.items[0], l.items[1:]...)
}

func MaxList[T cmp.Ordered](l *List[T]) T {
	if len(l.items) == 0 {
		panic(valueError("max() iterable argument is empty"))
	}
	return Max(l.items[0], l.items[1:]...)
}

func SumInt(l *List[int64]) int64 {
	var total int64
	for _, x := range l.items {
		total += x
	}
	return total
}

// SumFloat follows CPython 3.12+, whose sum() compensates for rounding
// error (Neumaier's variant of Kahan summation), so sum([0.1] * 10) is 1.0.
func SumFloat(l *List[float64]) float64 {
	var sum, c float64
	for _, x := range l.items {
		t := sum + x
		if math.Abs(sum) >= math.Abs(x) {
			c += (sum - t) + x
		} else {
			c += (x - t) + sum
		}
		sum = t
	}
	if c != 0 && !math.IsInf(c, 0) && !math.IsNaN(c) {
		sum += c
	}
	return sum
}

func DivModInt(a, b int64) Tuple2[int64, int64] {
	return Tup2(FloorDiv(a, b), Mod(a, b))
}

func DivModFloat(a, b float64) Tuple2[float64, float64] {
	if b == 0 {
		panic(zeroDivision("float divmod()"))
	}
	div, mod := divmodFloat(a, b)
	return Tup2(div, mod)
}

// PowInt is base ** exp for a non-negative exp on ints of a fixed size; a
// result beyond 64 bits is an OverflowError.
func PowInt(base, exp int64) int64 {
	r, ok := bigPow(big.NewInt(base), big.NewInt(exp)).(int64)
	if !ok {
		panic(overflowError("int too large for a 64-bit int"))
	}
	return r
}

// PowFloat is ** when either operand is a float, or an int raised to a
// negative power. Python raises where C's pow returns Inf, and returns a
// complex number for a negative base with a fractional exponent, which
// this runtime cannot represent.
func PowFloat(a, b float64) float64 {
	if a == 0 && b < 0 {
		panic(zeroDivision("0.0 cannot be raised to a negative power"))
	}
	if a < 0 && !math.IsInf(a, 0) && !math.IsInf(b, 0) && b != math.Trunc(b) {
		panic(valueError("negative number raised to a fractional power would be complex"))
	}
	r := math.Pow(a, b)
	if math.IsInf(r, 0) && !math.IsInf(a, 0) && !math.IsInf(b, 0) {
		panic(overflowError("(34, 'Numerical result out of range')"))
	}
	return r
}

// MinAny and MaxAny are min and max for items ordered by their own methods,
// such as instances of classes with __lt__. They keep the first of equal
// items, comparing the way Python does: min asks item < best, max item > best.
func MinAny[T any](first T, rest ...T) T {
	best := first
	for _, x := range rest {
		if Lt(x, best) {
			best = x
		}
	}
	return best
}

func MaxAny[T any](first T, rest ...T) T {
	best := first
	for _, x := range rest {
		if Gt(x, best) {
			best = x
		}
	}
	return best
}

func MinAnyList[T any](l *List[T]) T {
	if len(l.items) == 0 {
		panic(valueError("min() iterable argument is empty"))
	}
	return MinAny(l.items[0], l.items[1:]...)
}

func MaxAnyList[T any](l *List[T]) T {
	if len(l.items) == 0 {
		panic(valueError("max() iterable argument is empty"))
	}
	return MaxAny(l.items[0], l.items[1:]...)
}

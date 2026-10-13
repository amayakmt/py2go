package runtime

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strings"
)

// Ints beyond 64 bits are *big.Int. A value is always normalized: an
// int64 when it fits, a *big.Int otherwise, so equal ints compare as equal
// Go values and small arithmetic stays on machine words.

// MaxStrDigits is sys.get_int_max_str_digits(): the longest decimal string an
// int converts to or from.
const MaxStrDigits = 4300

var (
	bigOne     = big.NewInt(1)
	bigMinInt  = big.NewInt(math.MinInt64)
	bigMaxInt  = big.NewInt(math.MaxInt64)
	bigZero    = big.NewInt(0)
	bigTenTo18 = big.NewInt(1e18)
)

// NormBig returns the int64 when the value fits.
func NormBig(b *big.Int) any {
	if b.IsInt64() {
		return b.Int64()
	}
	return b
}

// BigLit is an int literal that does not fit 64 bits.
func BigLit(text string) any {
	b, ok := new(big.Int).SetString(text, 10)
	if !ok {
		panic(valueError("bad int literal " + text))
	}
	return NormBig(b)
}

// bigOf reads any int (int64, bool, *big.Int) as a *big.Int. The result of an
// int64 is a fresh value; the *big.Int is not copied.
func bigOf(v any) (*big.Int, bool) {
	switch x := v.(type) {
	case int64:
		return big.NewInt(x), true
	case int:
		return big.NewInt(int64(x)), true
	case bool:
		return big.NewInt(b2i(x)), true
	case *big.Int:
		return x, true
	}
	return nil, false
}

// isIntVal reports whether v is an int of any size (a bool counts).
func isIntVal(v any) bool {
	switch v.(type) {
	case int64, int, bool, *big.Int:
		return true
	}
	return false
}

func isBig(v any) bool { _, ok := v.(*big.Int); return ok }

// bigToFloat is float(x), an OverflowError when x is too large.
func bigToFloat(b *big.Int) float64 {
	f, _ := new(big.Float).SetInt(b).Float64()
	if math.IsInf(f, 0) {
		panic(overflowError("int too large to convert to float"))
	}
	return f
}

// BigFromFloat is int(f) for a float beyond 64 bits.
func bigFromFloat(f float64) any {
	b, _ := new(big.Float).SetFloat64(f).Int(nil)
	return NormBig(b)
}

// intArith applies a binary operator to two ints of any size. ok is false
// when the operator does not apply to ints (the caller reports the error).
func intArith(op string, a, b any) (any, bool) {
	if x, ok := a.(int64); ok {
		if y, ok := b.(int64); ok {
			if r, done := smallArith(op, x, y); done {
				return r, true
			}
		}
	}
	x, _ := bigOf(a)
	y, _ := bigOf(b)
	switch op {
	case "Add":
		return NormBig(new(big.Int).Add(x, y)), true
	case "Sub":
		return NormBig(new(big.Int).Sub(x, y)), true
	case "Mult":
		return NormBig(new(big.Int).Mul(x, y)), true
	case "Div":
		return bigTrueDiv(x, y), true
	case "FloorDiv":
		q, _ := bigFloorDivMod(x, y, "integer division or modulo by zero")
		return NormBig(q), true
	case "Mod":
		_, r := bigFloorDivMod(x, y, "integer modulo by zero")
		return NormBig(r), true
	case "Pow":
		return bigPow(x, y), true
	case "BitAnd":
		return NormBig(new(big.Int).And(x, y)), true
	case "BitOr":
		return NormBig(new(big.Int).Or(x, y)), true
	case "BitXor":
		return NormBig(new(big.Int).Xor(x, y)), true
	case "LShift":
		if y.Sign() < 0 {
			panic(valueError("negative shift count"))
		}
		if !y.IsInt64() || y.Int64() > 1<<32 {
			if x.Sign() == 0 {
				return int64(0), true
			}
			if !y.IsInt64() {
				panic(overflowError("too many digits in integer"))
			}
			panic(newError("MemoryError", ""))
		}
		return NormBig(new(big.Int).Lsh(x, uint(y.Int64()))), true
	case "RShift":
		if y.Sign() < 0 {
			panic(valueError("negative shift count"))
		}
		if !y.IsInt64() || y.Int64() > 1<<32 {
			if x.Sign() < 0 {
				return int64(-1), true
			}
			return int64(0), true
		}
		return NormBig(new(big.Int).Rsh(x, uint(y.Int64()))), true
	}
	return nil, false
}

// smallArith is the machine-word case. done is false when the result needs
// more than 64 bits or the operator is not handled here.
func smallArith(op string, x, y int64) (any, bool) {
	switch op {
	case "Add":
		c := x + y
		if (c > x) == (y > 0) {
			return c, true
		}
	case "Sub":
		c := x - y
		if (c < x) == (y > 0) {
			return c, true
		}
	case "Mult":
		hi, lo := bits.Mul64(uint64(abs64(x)), uint64(abs64(y)))
		if hi == 0 && lo <= math.MaxInt64 && x != math.MinInt64 && y != math.MinInt64 {
			if (x < 0) != (y < 0) {
				return -int64(lo), true
			}
			return int64(lo), true
		}
		if x == 0 || y == 0 {
			return int64(0), true
		}
	case "FloorDiv":
		if y != 0 && !(x == math.MinInt64 && y == -1) {
			return FloorDiv(x, y), true
		}
	case "Mod":
		if y != 0 {
			return Mod(x, y), true
		}
	case "Div":
		if y != 0 && x > -(1<<53) && x < 1<<53 && y > -(1<<53) && y < 1<<53 {
			return float64(x) / float64(y), true
		}
	case "BitAnd":
		return x & y, true
	case "BitOr":
		return x | y, true
	case "BitXor":
		return x ^ y, true
	case "RShift":
		if y >= 0 {
			if y > 63 {
				y = 63
			}
			return x >> uint(y), true
		}
	case "LShift":
		if y >= 0 && y < 63 {
			r := x << uint(y)
			if r>>uint(y) == x {
				return r, true
			}
		}
	}
	return nil, false
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// bigFloorDivMod is divmod for ints: the quotient rounds toward negative
// infinity and the remainder takes the sign of the divisor.
func bigFloorDivMod(x, y *big.Int, zeroMsg string) (*big.Int, *big.Int) {
	if y.Sign() == 0 {
		panic(zeroDivision(zeroMsg))
	}
	q, r := new(big.Int).QuoRem(x, y, new(big.Int))
	if r.Sign() != 0 && (r.Sign() < 0) != (y.Sign() < 0) {
		q.Sub(q, bigOne)
		r.Add(r, y)
	}
	return q, r
}

// bigTrueDiv is x / y rounded to the nearest float, as CPython does.
func bigTrueDiv(x, y *big.Int) float64 {
	if y.Sign() == 0 {
		panic(zeroDivision("division by zero"))
	}
	f, _ := new(big.Rat).SetFrac(x, y).Float64()
	if math.IsInf(f, 0) {
		panic(overflowError("integer division result too large for a float"))
	}
	return f
}

// bigPow is x ** y for ints: an int for a non-negative exponent, a float for a
// negative one.
func bigPow(x, y *big.Int) any {
	if y.Sign() < 0 {
		return PowFloat(bigToFloatLoose(x), bigToFloatLoose(y))
	}
	if x.IsInt64() && x.Int64() == 1 {
		return int64(1)
	}
	if x.IsInt64() && x.Int64() == 0 {
		if y.Sign() == 0 {
			return int64(1)
		}
		return int64(0)
	}
	if x.IsInt64() && x.Int64() == -1 {
		if y.Bit(0) == 0 {
			return int64(1)
		}
		return int64(-1)
	}
	if !y.IsInt64() || y.Int64() > 1<<27 {
		panic(overflowError("exponent too large"))
	}
	return NormBig(new(big.Int).Exp(x, y, nil))
}

func bigToFloatLoose(b *big.Int) float64 {
	f, _ := new(big.Float).SetInt(b).Float64()
	return f
}

// bigNegate, bigAbs and bigInvert are the unary operators.
func intUnary(op string, v any) (any, bool) {
	b, ok := bigOf(v)
	if !ok {
		return nil, false
	}
	switch op {
	case "USub":
		return NormBig(new(big.Int).Neg(b)), true
	case "UAdd":
		return NormBig(b), true
	case "Invert":
		return NormBig(new(big.Int).Not(b)), true
	case "Abs":
		return NormBig(new(big.Int).Abs(b)), true
	}
	return nil, false
}

// BigString is str(x) for an int, with the digit limit of CPython.
func BigString(b *big.Int) string {
	if b.BitLen() > 14284 { // about 4300 digits
		if n := len(b.String()); n > MaxStrDigits+1 || (n > MaxStrDigits && b.Sign() >= 0) {
			panic(valueError("Exceeds the limit (4300 digits) for integer string conversion; use sys.set_int_max_str_digits() to increase the limit"))
		}
	}
	return b.String()
}

// IntFromText is int(text, base) for any size, with the digit limit.
func IntFromText(s string, base int, orig string) any {
	t := asciiDigits(strings.ReplaceAll(s, "_", ""))
	digits := len(strings.TrimLeft(strings.TrimLeft(t, "+-"), "0"))
	if base == 10 && digits > MaxStrDigits {
		panic(valueError(fmt.Sprintf("Exceeds the limit (4300 digits) for integer string conversion: value has %d digits; use sys.set_int_max_str_digits() to increase the limit", digits)))
	}
	b, ok := new(big.Int).SetString(t, base)
	if !ok {
		panic(valueError(fmt.Sprintf("invalid literal for int() with base %d: %s", base, quote(orig))))
	}
	return NormBig(b)
}

// bigCompare orders two ints of any size.
func bigCompare(a, b any) int {
	x, _ := bigOf(a)
	y, _ := bigOf(b)
	return x.Cmp(y)
}

// bigFloatCompare orders an int against a float exactly (no rounding of the
// int), with NaN unordered: ok is false then.
func bigFloatCompare(b *big.Int, f float64) (int, bool) {
	if math.IsNaN(f) {
		return 0, false
	}
	if math.IsInf(f, 1) {
		return -1, true
	}
	if math.IsInf(f, -1) {
		return 1, true
	}
	return new(big.Float).SetInt(b).Cmp(new(big.Float).SetFloat64(f)), true
}

// IntBitLength is int.bit_length().
func IntBitLength(v any) int64 {
	b, _ := bigOf(v)
	return int64(b.BitLen())
}

// bigHash is hash(x) for an int: x modulo 2**61 - 1 with the sign kept.
func bigHash(b *big.Int) int64 {
	m := new(big.Int).Sub(new(big.Int).Lsh(bigOne, 61), bigOne)
	r := new(big.Int).Rem(new(big.Int).Abs(b), m).Int64()
	if b.Sign() < 0 {
		r = -r
	}
	if r == -1 {
		r = -2
	}
	return r
}

// asciiDigits turns the decimal digits of other scripts into 0 to 9, as
// int() and float() read them.
func asciiDigits(s string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		if r >= 0x80 {
			for _, rg := range decimalRanges {
				if r >= rg[0] && r <= rg[1] {
					r = '0' + (r-rg[0])%10
					break
				}
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

package runtime

import (
	"math"
	"math/big"
)

// Div is Python's / when either operand is a float. It always yields a
// float and raises on a zero divisor, where Go would return +Inf, -Inf or NaN.
func Div(a, b float64) float64 {
	if b == 0 {
		panic(zeroDivision("float division by zero"))
	}
	return a / b
}

// DivInt is Python's / on two ints, correctly rounded as CPython does.
func DivInt(a, b int64) float64 {
	if b == 0 {
		panic(zeroDivision("division by zero"))
	}
	if a > -(1<<53) && a < 1<<53 && b > -(1<<53) && b < 1<<53 {
		return float64(a) / float64(b)
	}
	return bigTrueDiv(big.NewInt(a), big.NewInt(b))
}

// MulInt is a * b on ints; a product beyond 64 bits is an OverflowError
// where the operands have a fixed size.
func MulInt(a, b int64) int64 {
	if r, ok := smallArith("Mult", a, b); ok {
		return r.(int64)
	}
	panic(overflowError("int too large for a 64-bit int"))
}

// FloorDiv is Python's // on ints: the quotient rounds toward negative
// infinity. Go's / truncates toward zero.
func FloorDiv(a, b int64) int64 {
	if b == 0 {
		panic(zeroDivision("integer division or modulo by zero"))
	}
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// Mod is Python's % on ints: the result takes the sign of the divisor.
// Go's % takes the sign of the dividend.
func Mod(a, b int64) int64 {
	if b == 0 {
		panic(zeroDivision("integer modulo by zero"))
	}
	r := a % b
	if r != 0 && (r < 0) != (b < 0) {
		r += b
	}
	return r
}

// divmodFloat follows CPython's float_divmod. Computing floor(a/b) directly
// can be off by one when a/b rounds up to an integer.
func divmodFloat(a, b float64) (div, mod float64) {
	mod = math.Mod(a, b)
	div = (a - mod) / b
	if mod != 0 {
		if (b < 0) != (mod < 0) {
			mod += b
			div--
		}
	} else {
		mod = math.Copysign(0, b)
	}
	if div != 0 {
		floordiv := math.Floor(div)
		if div-floordiv > 0.5 {
			floordiv++
		}
		return floordiv, mod
	}
	return math.Copysign(0, a/b), mod
}

// FloorDivFloat is Python's // when either operand is a float.
func FloorDivFloat(a, b float64) float64 {
	if b == 0 {
		panic(zeroDivision("float floor division by zero"))
	}
	div, _ := divmodFloat(a, b)
	return div
}

// ModFloat is Python's % when either operand is a float.
func ModFloat(a, b float64) float64 {
	if b == 0 {
		panic(zeroDivision("float modulo by zero"))
	}
	_, mod := divmodFloat(a, b)
	return mod
}

// LShift is Python's << on ints. A negative count raises ValueError.
// Bits shifted out of int64 are lost, a known limit of the int64 mapping.
func LShift(a, n int64) int64 {
	if n < 0 {
		panic(valueError("negative shift count"))
	}
	if a == 0 {
		return 0
	}
	if n >= 63 || (a<<uint(n))>>uint(n) != a {
		panic(overflowError("int too large for a 64-bit int"))
	}
	return a << uint(n)
}

// RShift is Python's >> on ints. A negative count raises ValueError.
func RShift(a, n int64) int64 {
	if n < 0 {
		panic(valueError("negative shift count"))
	}
	if n >= 64 {
		n = 63
	}
	return a >> uint(n)
}

// CheckRangeStep rejects a zero step the way range() does.
func CheckRangeStep(step int64) {
	if step == 0 {
		panic(valueError("range() arg 3 must not be zero"))
	}
}

// NegZero is -0.0. Go constant arithmetic has no negative zero,
// so a literal -0.0 cannot be written as a constant expression.
var NegZero = math.Copysign(0, -1)

// Inf is float('inf'). Go has no literal for it, and Python spells it
// 1e999 in source, which the decoder turns into +Inf.
var Inf = math.Inf(1)

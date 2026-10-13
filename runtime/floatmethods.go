package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

func floatMethod(f float64, name string, args []any) (any, bool) {
	switch name {
	case "is_integer":
		return !math.IsInf(f, 0) && !math.IsNaN(f) && f == math.Trunc(f), true
	case "conjugate", "__float__":
		return f, true
	case "hex":
		return FloatHex(f), true
	case "as_integer_ratio":
		return floatRatio(f), true
	case "__trunc__":
		return DynIntAny(f), true
	}
	return nil, false
}

// FloatHex is float.hex(): '0x1.8000000000000p+0'.
func FloatHex(f float64) string {
	switch {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	}
	sign := ""
	if math.Signbit(f) {
		sign, f = "-", -f
	}
	if f == 0 {
		return sign + "0x0.0p+0"
	}
	bitsOf := math.Float64bits(f)
	exp := int((bitsOf>>52)&0x7ff) - 1023
	mant := bitsOf & (1<<52 - 1)
	lead := "1"
	if exp == -1023 { // subnormal
		lead, exp = "0", -1022
	}
	return fmt.Sprintf("%s0x%s.%013xp%+d", sign, lead, mant, exp)
}

// FloatFromHex is float.fromhex(text).
func FloatFromHex(text string) float64 {
	t := strings.TrimSpace(strings.ToLower(text))
	switch strings.TrimLeft(t, "+-") {
	case "inf", "infinity":
		if strings.HasPrefix(t, "-") {
			return math.Inf(-1)
		}
		return math.Inf(1)
	case "nan":
		return math.NaN()
	}
	body := strings.TrimLeft(t, "+-")
	if !strings.HasPrefix(body, "0x") {
		t = t[:len(t)-len(body)] + "0x" + body
	}
	if !strings.Contains(t, "p") {
		t += "p0"
	}
	f, err := strconv.ParseFloat(t, 64)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
			panic(overflowError("hexadecimal value too large to represent as a float"))
		}
		panic(valueError("invalid hexadecimal floating-point string"))
	}
	return f
}

func floatRatio(f float64) any {
	switch {
	case math.IsNaN(f):
		panic(valueError("cannot convert NaN to integer ratio"))
	case math.IsInf(f, 0):
		panic(overflowError("cannot convert Infinity to integer ratio"))
	}
	r := new(big.Rat)
	r.SetFloat64(f)
	return NewVarTuple[any](NormBig(new(big.Int).Set(r.Num())), NormBig(new(big.Int).Set(r.Denom())))
}

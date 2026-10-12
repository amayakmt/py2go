package runtime

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Print is Python's print(*args) with the default sep and end.
func Print(args ...any) {
	var b strings.Builder
	for i, a := range args {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(Str(a))
	}
	b.WriteByte('\n')
	os.Stdout.WriteString(b.String())
}

// Str is Python's str() for the types the transpiler supports.
func Str(v any) string {
	switch v := v.(type) {
	case bool:
		if v {
			return "True"
		}
		return "False"
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case float64:
		return FloatRepr(v)
	case string:
		return v
	default:
		panic(fmt.Sprintf("runtime.Str: unsupported type %T", v))
	}
}

// FloatRepr formats a float the way Python's repr() does: the shortest
// string that round-trips, in scientific form when the decimal exponent is
// below -4 or at least 16.
func FloatRepr(f float64) string {
	switch {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	}

	// 'e' with precision -1 gives the shortest digits: "-d.ddde+XX".
	s := strconv.FormatFloat(f, 'e', -1, 64)
	sign := ""
	if s[0] == '-' {
		sign, s = "-", s[1:]
	}
	mant, expText, _ := strings.Cut(s, "e")
	exp, _ := strconv.Atoi(expText)
	digits := strings.Replace(mant, ".", "", 1)

	if exp < -4 || exp >= 16 {
		out := digits[:1]
		if len(digits) > 1 {
			out += "." + digits[1:]
		}
		esign := "+"
		if exp < 0 {
			esign, exp = "-", -exp
		}
		return fmt.Sprintf("%s%se%s%02d", sign, out, esign, exp)
	}
	if exp < 0 {
		return sign + "0." + strings.Repeat("0", -exp-1) + digits
	}
	if len(digits) <= exp+1 {
		return sign + digits + strings.Repeat("0", exp+1-len(digits)) + ".0"
	}
	return sign + digits[:exp+1] + "." + digits[exp+1:]
}

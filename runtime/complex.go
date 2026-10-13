package runtime

import (
	"fmt"
	"math"
	"math/cmplx"
	"strconv"
	"strings"
)

// complex numbers are complex128 values.

// complexFmt writes one part of a complex the way repr does: shortest digits,
// and no ".0" on an integral value.
func complexFmt(f float64) string {
	s := FloatRepr(f)
	return strings.TrimSuffix(s, ".0")
}

// ComplexRepr is repr(z).
func ComplexRepr(z complex128) string {
	re, im := real(z), imag(z)
	if re == 0 && !math.Signbit(re) {
		return complexFmt(im) + "j"
	}
	sign := "+"
	if math.Signbit(im) || math.IsNaN(im) && false {
		sign = ""
	}
	ims := complexFmt(im)
	if sign == "+" && strings.HasPrefix(ims, "-") {
		sign = ""
	}
	return "(" + complexFmt(re) + sign + ims + "j)"
}

func isComplexVal(v any) bool { _, ok := v.(complex128); return ok }

// toComplex reads a number as a complex.
func toComplex(v any) (complex128, bool) {
	switch x := v.(type) {
	case complex128:
		return x, true
	case float64:
		return complex(x, 0), true
	}
	if b, ok := bigOf(v); ok {
		return complex(bigToFloat(b), 0), true
	}
	return 0, false
}

func complexBinary(op string, a, b complex128) any {
	switch op {
	case "Add":
		return a + b
	case "Sub":
		return a - b
	case "Mult":
		return a * b
	case "Div":
		if b == 0 {
			panic(zeroDivision("complex division by zero"))
		}
		return complexDiv(a, b)
	case "Pow":
		return complexPow(a, b)
	}
	return nil
}

// complexDiv follows CPython's algorithm (Smith's method), which keeps results
// that a naive formula would lose to overflow.
func complexDiv(a, b complex128) complex128 {
	abs_breal, abs_bimag := math.Abs(real(b)), math.Abs(imag(b))
	var re, im float64
	if abs_breal >= abs_bimag {
		if abs_breal == 0 {
			panic(zeroDivision("complex division by zero"))
		}
		ratio := imag(b) / real(b)
		denom := real(b) + imag(b)*ratio
		re = (real(a) + imag(a)*ratio) / denom
		im = (imag(a) - real(a)*ratio) / denom
	} else {
		ratio := real(b) / imag(b)
		denom := real(b)*ratio + imag(b)
		re = (real(a)*ratio + imag(a)) / denom
		im = (imag(a)*ratio - real(a)) / denom
	}
	return complex(re, im)
}

func complexPow(a, b complex128) complex128 {
	if b == 0 {
		return 1
	}
	if a == 0 {
		if imag(b) != 0 || real(b) < 0 {
			panic(zeroDivision("0.0 to a negative or complex power"))
		}
		return 0
	}
	if imag(b) == 0 && real(b) == math.Trunc(real(b)) && math.Abs(real(b)) <= 100 {
		n := int(real(b))
		neg := n < 0
		if neg {
			n = -n
		}
		r := complex(1, 0)
		for p, e := a, n; e > 0; e >>= 1 {
			if e&1 == 1 {
				r *= p
			}
			p *= p
		}
		if neg {
			return complexDiv(1, r)
		}
		return r
	}
	// CPython's formula, which rounds the way CPython does.
	vabs := math.Hypot(real(a), imag(a))
	length := math.Pow(vabs, real(b))
	at := math.Atan2(imag(a), real(a))
	phase := at * real(b)
	if imag(b) != 0 {
		length /= math.Exp(at * imag(b))
		phase += imag(b) * math.Log(vabs)
	}
	r := complex(length*math.Cos(phase), length*math.Sin(phase))
	if cmplx.IsInf(r) && !cmplx.IsInf(a) && !cmplx.IsInf(b) {
		panic(overflowError("complex exponentiation"))
	}
	return r
}

func complexHash(z complex128) int64 {
	h := hashFloat(real(z)) + 1000003*hashFloat(imag(z))
	return fixHash(h)
}

// ParseComplex is complex("1+2j").
func ParseComplex(s string) complex128 {
	t := strings.TrimSpace(s)
	bad := func() complex128 { panic(valueError("complex() arg is a malformed string")) }
	t = strings.TrimSuffix(strings.TrimPrefix(t, "("), ")")
	if t == "" {
		return bad()
	}
	if !strings.HasSuffix(t, "j") && !strings.HasSuffix(t, "J") {
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return bad()
		}
		return complex(f, 0)
	}
	t = t[:len(t)-1]
	// Split at the last sign that is not part of an exponent.
	split := -1
	for i := len(t) - 1; i > 0; i-- {
		if (t[i] == '+' || t[i] == '-') && t[i-1] != 'e' && t[i-1] != 'E' {
			split = i
			break
		}
	}
	parse := func(x string) (float64, bool) {
		switch x {
		case "", "+":
			return 1, true
		case "-":
			return -1, true
		}
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	}
	if split < 0 {
		im, ok := parse(t)
		if !ok {
			return bad()
		}
		return complex(0, im)
	}
	re, ok1 := strconv.ParseFloat(t[:split], 64)
	im, ok2 := parse(t[split:])
	if ok1 != nil || !ok2 {
		return bad()
	}
	return complex(re, im)
}

// NewComplex is complex(real=0, imag=0).
func NewComplex(args []any, kw []KW) complex128 {
	var re, im any = int64(0), int64(0)
	if len(args) > 0 {
		re = Unbox(args[0])
	}
	if len(args) > 1 {
		im = Unbox(args[1])
	}
	if v, ok := kwValue(kw, "real"); ok {
		re = Unbox(v)
	}
	if v, ok := kwValue(kw, "imag"); ok {
		im = Unbox(v)
	}
	if s, ok := re.(string); ok {
		if len(args) > 1 {
			panic(newError("TypeError", "complex() can't take second arg if first is a string"))
		}
		return ParseComplex(s)
	}
	if _, ok := im.(string); ok {
		panic(newError("TypeError", "complex() second arg can't be a string"))
	}
	if !isComplexVal(re) && !isComplexVal(im) {
		if fr, ok := realOf(re); ok {
			if fi, ok := realOf(im); ok {
				return complex(fr, fi)
			}
		}
	}
	a, ok1 := toComplex(re)
	b, ok2 := toComplex(im)
	if !ok1 {
		panic(newError("TypeError", "complex() first argument must be a string or a number, not '"+TypeName(re)+"'"))
	}
	if !ok2 {
		panic(newError("TypeError", "complex() second argument must be a number, not '"+TypeName(im)+"'"))
	}
	// real + imag * 1j, part by part so that signed zeros survive.
	return complex(real(a)-imag(b), imag(a)+real(b))
}

func complexAttr(z complex128, name string) (any, bool) {
	switch name {
	case "real":
		return real(z), true
	case "imag":
		return imag(z), true
	case "conjugate":
		return func(args ...any) any { return cmplx.Conj(z) }, true
	}
	return nil, false
}

var _ = fmt.Sprint

// realOf reads an int, a bool or a float as a float.
func realOf(v any) (float64, bool) {
	if f, ok := v.(float64); ok {
		return f, true
	}
	if b, ok := bigOf(v); ok {
		return bigToFloat(b), true
	}
	return 0, false
}

// FormatComplex is format(z, spec). The real and imaginary parts are written
// with the spec's type and precision; without a type the text is the repr's,
// with the real part and the parentheses left out when it is a plain zero.
func FormatComplex(z complex128, s string) string {
	sp, err := parseSpec(s)
	if err != nil {
		panic(valueError(err.Error()))
	}
	switch {
	case sp.align == '=':
		panic(valueError("'=' alignment flag is not allowed in complex format specifier"))
	case sp.zero:
		panic(valueError("Zero padding is not allowed in complex format specifier"))
	case sp.typ != 0 && !strings.ContainsRune("eEfFgGn", rune(sp.typ)):
		panic(valueError("Unknown format code '" + string(sp.typ) + "' for object of type 'complex'"))
	}
	part := func(f float64, sign byte) string {
		sub := string(sign)
		if sp.alt {
			sub += "#"
		}
		if sp.hasPrec {
			sub += "." + strconv.Itoa(sp.prec)
		}
		t := sp.typ
		if t == 0 && sp.hasPrec {
			t = 'g'
		}
		if t != 0 {
			sub += string(t)
		}
		out := FormatFloat(f, sub)
		if sp.typ == 0 && !sp.hasPrec {
			out = strings.TrimSuffix(out, ".0")
		}
		return out
	}
	re, im := real(z), imag(z)
	var body string
	if sp.typ == 0 && re == 0 && !math.Signbit(re) {
		body = part(im, sp.sign) + "j"
	} else {
		body = part(re, sp.sign) + part(im, '+') + "j"
		if sp.typ == 0 {
			body = "(" + body + ")"
		}
	}
	sp.zero = false
	return sp.pad('>', "", body)
}

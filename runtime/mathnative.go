package runtime

import (
	"math"
	"math/big"
)

// The native part of the math module. Go's math package and the platform
// libm can disagree in the last digit of exp, log, pow and the
// trigonometric functions; the results are otherwise CPython's, down to the
// ValueError and OverflowError cases.

func domainError() *PyError { return valueError("math domain error") }
func rangeError() *PyError  { return newError("OverflowError", "math range error") }

// mathFloat reads the argument of a math function as a float.
func mathFloat(v any) float64 {
	v = Unbox(v)
	switch x := v.(type) {
	case float64:
		return x
	case bool:
		return float64(b2i(x))
	}
	if b, ok := bigOf(v); ok {
		return bigToFloat(b)
	}
	if n, ok := v.(Named); ok && DynHasAttr(n, "__float__") {
		return DynFloat(DynCall(DynGetAttr(n, "__float__")))
	}
	panic(newError("TypeError", "must be real number, not "+TypeName(v)))
}

// result checks a float result: an infinity from finite arguments is a range
// error, a nan from non-nan arguments a domain error.
func mathResult(r float64, args ...float64) float64 {
	if math.IsNaN(r) {
		for _, a := range args {
			if math.IsNaN(a) {
				return r
			}
		}
		panic(domainError())
	}
	if math.IsInf(r, 0) {
		for _, a := range args {
			if math.IsInf(a, 0) {
				return r
			}
		}
		panic(rangeError())
	}
	return r
}

func unary(name string, f func(float64) float64) {
	RegisterNative("math."+name, func(args []any, kw []KW) any {
		nativeArgs(name, args, 1, 1)
		x := mathFloat(args[0])
		return mathResult(f(x), x)
	})
}

// intLike reads an argument that must be an int (a bool is one).
func intLike(v any) *big.Int {
	b, ok := bigOf(Unbox(v))
	if !ok {
		panic(newError("TypeError", "'"+TypeName(Unbox(v))+"' object cannot be interpreted as an integer"))
	}
	return b
}

func floatToInt(f float64) any {
	switch {
	case math.IsNaN(f):
		panic(valueError("cannot convert float NaN to integer"))
	case math.IsInf(f, 0):
		panic(newError("OverflowError", "cannot convert float infinity to integer"))
	}
	if f >= -(1<<62) && f <= 1<<62 {
		return int64(f)
	}
	return bigFromFloat(f)
}

func init() {
	unary("sqrt", func(x float64) float64 {
		if x < 0 {
			panic(domainError())
		}
		return math.Sqrt(x)
	})
	unary("cbrt", glibcCbrt)
	unary("exp", CrExp)
	unary("exp2", CrExp2)
	unary("expm1", math.Expm1)
	unary("sin", func(x float64) float64 {
		if math.IsInf(x, 0) {
			panic(domainError())
		}
		return CrSin(x)
	})
	unary("cos", func(x float64) float64 {
		if math.IsInf(x, 0) {
			panic(domainError())
		}
		return CrCos(x)
	})
	unary("tan", func(x float64) float64 {
		if math.IsInf(x, 0) {
			panic(domainError())
		}
		return CrTan(x)
	})
	unary("asin", CrAsin)
	unary("acos", CrAcos)
	unary("atan", CrAtan)
	unary("sinh", glibcSinh)
	unary("cosh", glibcCosh)
	unary("tanh", glibcTanh)
	unary("asinh", CrAsinh)
	unary("acosh", func(x float64) float64 {
		if x < 1 {
			panic(domainError())
		}
		return CrAcosh(x)
	})
	unary("atanh", func(x float64) float64 {
		if x <= -1 || x >= 1 {
			panic(domainError())
		}
		return CrAtanh(x)
	})
	unary("fabs", math.Abs)
	unary("degrees", func(x float64) float64 { return x * (180 / math.Pi) })
	unary("radians", func(x float64) float64 { return x * (math.Pi / 180) })
	unary("log1p", func(x float64) float64 {
		if x <= -1 {
			panic(domainError())
		}
		return math.Log1p(x)
	})
	unary("erf", glibcErf)
	unary("erfc", glibcErfc)
	unary("gamma", func(x float64) float64 {
		r, e := pyGamma(x)
		switch e {
		case errDomain:
			panic(domainError())
		case errRange:
			panic(rangeError())
		}
		return r
	})
	unary("lgamma", func(x float64) float64 {
		r, e := pyLgamma(x)
		switch e {
		case errDomain:
			panic(domainError())
		case errRange:
			panic(rangeError())
		}
		return r
	})

	RegisterNative("math.atan2", func(args []any, kw []KW) any {
		nativeArgs("atan2", args, 2, 2)
		return CrAtan2(mathFloat(args[0]), mathFloat(args[1]))
	})
	RegisterNative("math.copysign", func(args []any, kw []KW) any {
		nativeArgs("copysign", args, 2, 2)
		return math.Copysign(mathFloat(args[0]), mathFloat(args[1]))
	})
	RegisterNative("math.fmod", func(args []any, kw []KW) any {
		nativeArgs("fmod", args, 2, 2)
		x, y := mathFloat(args[0]), mathFloat(args[1])
		if math.IsInf(x, 0) && !math.IsNaN(y) || y == 0 && !math.IsNaN(x) {
			panic(domainError())
		}
		return math.Mod(x, y)
	})
	RegisterNative("math.remainder", func(args []any, kw []KW) any {
		nativeArgs("remainder", args, 2, 2)
		x, y := mathFloat(args[0]), mathFloat(args[1])
		if math.IsInf(x, 0) && !math.IsNaN(y) || y == 0 && !math.IsNaN(x) {
			panic(domainError())
		}
		return math.Remainder(x, y)
	})
	RegisterNative("math.pow", func(args []any, kw []KW) any {
		nativeArgs("pow", args, 2, 2)
		x, y := mathFloat(args[0]), mathFloat(args[1])
		if x == 0 && y < 0 && !math.IsInf(y, 0) && !math.IsNaN(y) {
			panic(domainError())
		}
		if x < 0 && !math.IsInf(x, 0) && y != math.Floor(y) && !math.IsInf(y, 0) && !math.IsNaN(y) {
			panic(domainError())
		}
		return mathResult(CrPow(x, y), x, y)
	})
	RegisterNative("math.hypot", func(args []any, kw []KW) any {
		fs := make([]float64, len(args))
		for i, a := range args {
			fs[i] = mathFloat(a)
		}
		return hypotAll(fs)
	})
	RegisterNative("math.dist", func(args []any, kw []KW) any {
		nativeArgs("dist", args, 2, 2)
		p, q := floatSeq(args[0]), floatSeq(args[1])
		if len(p) != len(q) {
			panic(valueError("both points must have the same number of dimensions"))
		}
		d := make([]float64, len(p))
		for i := range p {
			d[i] = p[i] - q[i]
		}
		return hypotAll(d)
	})
	RegisterNative("math.isclose", func(args []any, kw []KW) any {
		nativeArgs("isclose", args, 2, 2)
		a, b := mathFloat(args[0]), mathFloat(args[1])
		rel, abs := 1e-09, 0.0
		if v, ok := kwValue(kw, "rel_tol"); ok {
			rel = mathFloat(v)
		}
		if v, ok := kwValue(kw, "abs_tol"); ok {
			abs = mathFloat(v)
		}
		if rel < 0 || abs < 0 {
			panic(valueError("tolerances must be non-negative"))
		}
		if a == b {
			return true
		}
		if math.IsInf(a, 0) || math.IsInf(b, 0) {
			return false
		}
		diff := math.Abs(b - a)
		return diff <= math.Abs(rel*b) || diff <= math.Abs(rel*a) || diff <= abs
	})
	RegisterNative("math.isfinite", func(args []any, kw []KW) any {
		nativeArgs("isfinite", args, 1, 1)
		x := mathFloat(args[0])
		return !math.IsInf(x, 0) && !math.IsNaN(x)
	})
	RegisterNative("math.isinf", func(args []any, kw []KW) any {
		nativeArgs("isinf", args, 1, 1)
		return math.IsInf(mathFloat(args[0]), 0)
	})
	RegisterNative("math.isnan", func(args []any, kw []KW) any {
		nativeArgs("isnan", args, 1, 1)
		return math.IsNaN(mathFloat(args[0]))
	})
	RegisterNative("math.ldexp", func(args []any, kw []KW) any {
		nativeArgs("ldexp", args, 2, 2)
		x := mathFloat(args[0])
		e := intLike(args[1])
		if !e.IsInt64() || e.Int64() > 1<<20 || e.Int64() < -(1<<20) {
			if e.Sign() < 0 || x == 0 || math.IsInf(x, 0) || math.IsNaN(x) {
				return math.Ldexp(x, -(1 << 20))
			}
			panic(rangeError())
		}
		return mathResult(math.Ldexp(x, int(e.Int64())), x)
	})
	RegisterNative("math.frexp", func(args []any, kw []KW) any {
		nativeArgs("frexp", args, 1, 1)
		m, e := math.Frexp(mathFloat(args[0]))
		return NewVarTuple[any](m, int64(e))
	})
	RegisterNative("math.modf", func(args []any, kw []KW) any {
		nativeArgs("modf", args, 1, 1)
		x := mathFloat(args[0])
		if math.IsInf(x, 0) {
			return NewVarTuple[any](math.Copysign(0, x), x)
		}
		ip, fp := math.Modf(x)
		return NewVarTuple[any](fp, ip)
	})
	RegisterNative("math.floor", func(args []any, kw []KW) any {
		nativeArgs("floor", args, 1, 1)
		if b, ok := bigOf(Unbox(args[0])); ok {
			return NormBig(b)
		}
		if n, ok := Unbox(args[0]).(Named); ok && DynHasAttr(n, "__floor__") {
			return DynCall(DynGetAttr(n, "__floor__"))
		}
		return floatToInt(math.Floor(mathFloat(args[0])))
	})
	RegisterNative("math.ceil", func(args []any, kw []KW) any {
		nativeArgs("ceil", args, 1, 1)
		if b, ok := bigOf(Unbox(args[0])); ok {
			return NormBig(b)
		}
		if n, ok := Unbox(args[0]).(Named); ok && DynHasAttr(n, "__ceil__") {
			return DynCall(DynGetAttr(n, "__ceil__"))
		}
		return floatToInt(math.Ceil(mathFloat(args[0])))
	})
	RegisterNative("math.trunc", func(args []any, kw []KW) any {
		nativeArgs("trunc", args, 1, 1)
		if b, ok := bigOf(Unbox(args[0])); ok {
			return NormBig(b)
		}
		if n, ok := Unbox(args[0]).(Named); ok && DynHasAttr(n, "__trunc__") {
			return DynCall(DynGetAttr(n, "__trunc__"))
		}
		return floatToInt(math.Trunc(mathFloat(args[0])))
	})
	logFn := func(name string, f func(float64) float64) {
		RegisterNative("math."+name, func(args []any, kw []KW) any {
			nativeArgs(name, args, 1, 1)
			return logOf(args[0], f)
		})
	}
	logFn("log10", glibcLog10)
	logFn("log2", CrLog2)
	RegisterNative("math.log", func(args []any, kw []KW) any {
		nativeArgs("log", args, 1, 2)
		x := logOf(args[0], CrLog)
		if len(args) == 1 {
			return x
		}
		base := logOf(args[1], CrLog)
		if base == 0 {
			panic(newError("ZeroDivisionError", "float division by zero"))
		}
		return x / base
	})
	RegisterNative("math.factorial", func(args []any, kw []KW) any {
		nativeArgs("factorial", args, 1, 1)
		n := intLike(args[0])
		if n.Sign() < 0 {
			panic(valueError("factorial() not defined for negative values"))
		}
		if !n.IsInt64() || n.Int64() > 1<<40 {
			panic(newError("OverflowError", "factorial() argument should not exceed 9223372036854775807"))
		}
		return NormBig(new(big.Int).MulRange(1, n.Int64()))
	})
	RegisterNative("math.comb", func(args []any, kw []KW) any {
		nativeArgs("comb", args, 2, 2)
		n, k := intLike(args[0]), intLike(args[1])
		if n.Sign() < 0 {
			panic(valueError("n must be a non-negative integer"))
		}
		if k.Sign() < 0 {
			panic(valueError("k must be a non-negative integer"))
		}
		if k.Cmp(n) > 0 {
			return int64(0)
		}
		if !k.IsInt64() {
			panic(newError("OverflowError", "min(n - k, k) must not exceed 9223372036854775807"))
		}
		return NormBig(new(big.Int).Binomial(i64max(n), k.Int64()))
	})
	RegisterNative("math.perm", func(args []any, kw []KW) any {
		nativeArgs("perm", args, 1, 2)
		n := intLike(args[0])
		if (len(args) == 1 || IsNone(args[1])) && n.Sign() < 0 {
			panic(valueError("factorial() not defined for negative values"))
		}
		if n.Sign() < 0 {
			panic(valueError("n must be a non-negative integer"))
		}
		if len(args) == 1 || IsNone(args[1]) {
			return NormBig(new(big.Int).MulRange(1, i64max(n)))
		}
		k := intLike(args[1])
		if k.Sign() < 0 {
			panic(valueError("k must be a non-negative integer"))
		}
		if k.Cmp(n) > 0 {
			return int64(0)
		}
		return NormBig(new(big.Int).MulRange(i64max(n)-k.Int64()+1, i64max(n)))
	})
	RegisterNative("math.gcd", func(args []any, kw []KW) any {
		g := new(big.Int)
		for _, a := range args {
			g.GCD(nil, nil, new(big.Int).Abs(g), new(big.Int).Abs(intLike(a)))
		}
		return NormBig(g)
	})
	RegisterNative("math.lcm", func(args []any, kw []KW) any {
		l := big.NewInt(1)
		for _, a := range args {
			x := new(big.Int).Abs(intLike(a))
			if x.Sign() == 0 {
				return int64(0)
			}
			g := new(big.Int).GCD(nil, nil, l, x)
			l.Mul(l, x.Quo(x, g))
		}
		return NormBig(l)
	})
	RegisterNative("math.isqrt", func(args []any, kw []KW) any {
		nativeArgs("isqrt", args, 1, 1)
		n := intLike(args[0])
		if n.Sign() < 0 {
			panic(valueError("isqrt() argument must be nonnegative"))
		}
		return NormBig(new(big.Int).Sqrt(n))
	})
	RegisterNative("math.prod", func(args []any, kw []KW) any {
		nativeArgs("prod", args, 1, 1)
		var acc any = int64(1)
		if v, ok := kwValue(kw, "start"); ok {
			acc = v
		}
		for c := DynIter(args[0]).Iter(); c.Next(); {
			acc = DynBinary("Mult", acc, c.Value())
		}
		return acc
	})
	RegisterNative("math.fsum", func(args []any, kw []KW) any {
		nativeArgs("fsum", args, 1, 1)
		var xs []float64
		for c := DynIter(args[0]).Iter(); c.Next(); {
			xs = append(xs, mathFloat(c.Value()))
		}
		return fsum(xs)
	})
	RegisterNative("math.sumprod", func(args []any, kw []KW) any {
		nativeArgs("sumprod", args, 2, 2)
		p, q := DynList(args[0]), DynList(args[1])
		if p.Len() != q.Len() {
			panic(valueError("Inputs are not the same length"))
		}
		var acc any = int64(0)
		for i := 0; i < int(p.Len()); i++ {
			acc = DynBinary("Add", acc, DynBinary("Mult", p.items[i], q.items[i]))
		}
		return acc
	})
	RegisterNative("math.nextafter", func(args []any, kw []KW) any {
		nativeArgs("nextafter", args, 2, 2)
		return math.Nextafter(mathFloat(args[0]), mathFloat(args[1]))
	})
	RegisterNative("math.ulp", func(args []any, kw []KW) any {
		nativeArgs("ulp", args, 1, 1)
		x := math.Abs(mathFloat(args[0]))
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return x
		}
		if x == math.MaxFloat64 {
			return x - math.Nextafter(x, 0)
		}
		return math.Nextafter(x, math.Inf(1)) - x
	})
	RegisterNative("math.fma", func(args []any, kw []KW) any {
		nativeArgs("fma", args, 3, 3)
		x, y, z := mathFloat(args[0]), mathFloat(args[1]), mathFloat(args[2])
		return mathResult(math.FMA(x, y, z), x, y, z)
	})
}

func i64max(b *big.Int) int64 {
	if b.IsInt64() {
		return b.Int64()
	}
	return math.MaxInt64
}

// logOf is the logarithm of an int or float. An int too large for a float
// is split as CPython does: f(m) + f(2) * e for m * 2**e.
func logOf(v any, f func(float64) float64) float64 {
	v = Unbox(v)
	if b, ok := bigOf(v); ok {
		if b.Sign() <= 0 {
			panic(domainError())
		}
		if f64 := bigToFloatLoose(b); !math.IsInf(f64, 0) {
			return f(f64)
		}
		m := new(big.Float)
		e := new(big.Float).SetInt(b).MantExp(m)
		mf, _ := m.Float64()
		return f(mf) + f(2.0)*float64(e)
	}
	x := mathFloat(v)
	if x <= 0 {
		if x == 0 || !math.IsNaN(x) {
			panic(domainError())
		}
	}
	return f(x)
}

func floatSeq(v any) []float64 {
	var out []float64
	for c := DynIter(v).Iter(); c.Next(); {
		out = append(out, mathFloat(c.Value()))
	}
	return out
}

// hypotAll is the Euclidean norm, scaled so that it neither overflows nor
// underflows.
func hypotAll(xs []float64) float64 {
	max := 0.0
	hasNaN := false
	for _, x := range xs {
		x = math.Abs(x)
		if math.IsInf(x, 0) {
			return math.Inf(1)
		}
		if math.IsNaN(x) {
			hasNaN = true
		}
		if x > max {
			max = x
		}
	}
	if hasNaN {
		return math.NaN()
	}
	if max == 0 || len(xs) == 1 {
		return max
	}
	abs := make([]float64, len(xs))
	for i, x := range xs {
		abs[i] = math.Abs(x)
	}
	return vectorNorm(abs, max)
}

// fsum is the exactly rounded sum of floats (Shewchuk's algorithm, as in
// CPython).
func fsum(xs []float64) float64 {
	var partials []float64
	special := 0.0
	for _, x := range xs {
		if math.IsInf(x, 0) || math.IsNaN(x) {
			special += x
			continue
		}
		i := 0
		for _, y := range partials {
			if math.Abs(x) < math.Abs(y) {
				x, y = y, x
			}
			hi := x + y
			lo := y - (hi - x)
			if lo != 0 {
				partials[i] = lo
				i++
			}
			x = hi
		}
		partials = append(partials[:i], x)
	}
	if special != 0 || math.IsNaN(special) {
		return special
	}
	for _, x := range xs {
		if math.IsInf(x, 0) {
			return special
		}
	}
	n := len(partials)
	if n == 0 {
		return 0
	}
	hi := partials[n-1]
	n--
	var lo float64
	for n > 0 {
		x := hi
		y := partials[n-1]
		n--
		hi = x + y
		lo = y - (hi - x)
		if lo != 0 {
			break
		}
	}
	if n > 0 && (lo < 0 && partials[n-1] < 0 || lo > 0 && partials[n-1] > 0) {
		y := lo * 2
		x := hi + y
		if y == x-hi {
			hi = x
		}
	}
	return hi
}

package runtime

import (
	"math"
	"math/cmplx"
)

// The native part of the cmath module, on Go's math/cmplx. Go follows C99's
// special values for infinities and nans, as CPython does; the results can
// differ from CPython's in the last digit. The errors are CPython's: a nan
// from arguments without one is a ValueError ("math domain error"), an
// infinity from finite arguments an OverflowError ("math range error").

// cmathArg reads the argument of a cmath function as a complex.
func cmathArg(v any) complex128 {
	v = BaseValue(Unbox(v))
	if z, ok := toComplex(v); ok {
		return z
	}
	if b, ok := v.(bool); ok {
		return complex(float64(b2i(b)), 0)
	}
	if n, ok := v.(Named); ok {
		if DynHasAttr(n, "__complex__") {
			r := Unbox(DynCall(DynGetAttr(n, "__complex__")))
			if z, ok := r.(complex128); ok {
				return z
			}
			panic(newError("TypeError", "__complex__ returned non-complex (type "+TypeName(r)+")"))
		}
		if DynHasAttr(n, "__float__") || DynHasAttr(n, "__index__") {
			return complex(mathFloat(n), 0)
		}
	}
	panic(newError("TypeError", "must be real number, not "+TypeName(v)))
}

func finiteC(z complex128) bool {
	return !math.IsInf(real(z), 0) && !math.IsInf(imag(z), 0) && !math.IsNaN(real(z)) && !math.IsNaN(imag(z))
}

func hasNaNC(z complex128) bool { return math.IsNaN(real(z)) || math.IsNaN(imag(z)) }

// cmathResult checks a result as CPython's errno tests do.
func cmathResult(r, z complex128) complex128 {
	if !finiteC(z) {
		return r
	}
	if math.IsInf(real(r), 0) || math.IsInf(imag(r), 0) {
		// An overflow, even where Go's formula then multiplies the
		// infinity by zero into a nan (exp(1000+0j)).
		panic(rangeError())
	}
	if hasNaNC(r) {
		panic(domainError())
	}
	return r
}

func init() {
	unaryC := func(name string, f func(complex128) complex128) {
		RegisterNative("cmath."+name, func(args []any, kw []KW) any {
			nativeArgs(name, args, 1, 1)
			z := cmathArg(args[0])
			return cmathResult(f(z), z)
		})
	}
	unaryC("exp", cmplx.Exp)
	unaryC("sqrt", cmplx.Sqrt)
	unaryC("sin", cmplx.Sin)
	unaryC("cos", cmplx.Cos)
	unaryC("tan", cmplx.Tan)
	unaryC("asin", cmplx.Asin)
	unaryC("acos", cmplx.Acos)
	unaryC("atan", cmplx.Atan)
	unaryC("sinh", cmplx.Sinh)
	unaryC("cosh", cmplx.Cosh)
	unaryC("tanh", cmplx.Tanh)
	unaryC("asinh", cmplx.Asinh)
	unaryC("acosh", cmplx.Acosh)
	unaryC("atanh", func(z complex128) complex128 {
		if z == 1 || z == -1 {
			panic(domainError())
		}
		return cmplx.Atanh(z)
	})
	logOf := func(z complex128) complex128 {
		if z == 0 {
			panic(domainError())
		}
		return cmplx.Log(z)
	}
	RegisterNative("cmath.log", func(args []any, kw []KW) any {
		nativeArgs("log", args, 1, 2)
		z := cmathArg(args[0])
		r := cmathResult(logOf(z), z)
		if len(args) == 2 {
			b := cmathArg(args[1])
			lb := cmathResult(logOf(b), b)
			if lb == 0 {
				panic(newError("ZeroDivisionError", "complex division by zero"))
			}
			return complexDiv(r, lb)
		}
		return r
	})
	RegisterNative("cmath.log10", func(args []any, kw []KW) any {
		nativeArgs("log10", args, 1, 1)
		z := cmathArg(args[0])
		if z == 0 {
			panic(domainError())
		}
		return cmathResult(cmplx.Log10(z), z)
	})
	RegisterNative("cmath.phase", func(args []any, kw []KW) any {
		nativeArgs("phase", args, 1, 1)
		return cmplx.Phase(cmathArg(args[0]))
	})
	RegisterNative("cmath.polar", func(args []any, kw []KW) any {
		nativeArgs("polar", args, 1, 1)
		z := cmathArg(args[0])
		r := math.Hypot(real(z), imag(z))
		if math.IsInf(r, 0) && finiteC(z) {
			panic(rangeError())
		}
		return NewVarTuple[any](r, cmplx.Phase(z))
	})
	RegisterNative("cmath.rect", func(args []any, kw []KW) any {
		nativeArgs("rect", args, 2, 2)
		r, phi := mathFloat(args[0]), mathFloat(args[1])
		if math.IsInf(phi, 0) && !math.IsNaN(r) && r != 0 {
			panic(domainError())
		}
		if r == 0 && !math.IsNaN(phi) && !math.IsInf(phi, 0) {
			// rect(0, phi) keeps the signs of cos and sin.
			return complex(r*math.Cos(phi), r*math.Sin(phi))
		}
		return cmplx.Rect(r, phi)
	})
	RegisterNative("cmath.isfinite", func(args []any, kw []KW) any {
		nativeArgs("isfinite", args, 1, 1)
		return finiteC(cmathArg(args[0]))
	})
	RegisterNative("cmath.isnan", func(args []any, kw []KW) any {
		nativeArgs("isnan", args, 1, 1)
		return hasNaNC(cmathArg(args[0]))
	})
	RegisterNative("cmath.isinf", func(args []any, kw []KW) any {
		nativeArgs("isinf", args, 1, 1)
		z := cmathArg(args[0])
		return math.IsInf(real(z), 0) || math.IsInf(imag(z), 0)
	})
	RegisterNative("cmath.isclose", func(args []any, kw []KW) any {
		nativeArgs("isclose", args, 4, 4)
		a, b := cmathArg(args[0]), cmathArg(args[1])
		rel, abs := mathFloat(args[2]), mathFloat(args[3])
		if rel < 0 || abs < 0 {
			panic(valueError("tolerances must be non-negative"))
		}
		if a == b {
			return true
		}
		if cmplx.IsInf(a) || cmplx.IsInf(b) {
			return false
		}
		diff := cmplx.Abs(a - b)
		return diff <= rel*cmplx.Abs(b) || diff <= rel*cmplx.Abs(a) || diff <= abs
	})
}

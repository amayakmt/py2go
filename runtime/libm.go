package runtime

import (
	"math"
	"math/big"
	"sync"
)

// Correctly rounded elementary functions. Go's math package is accurate to a
// unit in the last place or so, and the C library CPython calls rounds most of
// these functions correctly, so the last digit of repr() can differ. The
// functions here compute the result with 128 bits and round it once, which
// agrees with a correctly rounding libm (glibc's exp, log, pow, sin, cos and
// the like) except in the cases that no 128-bit computation can settle.

const crPrec = 128

func nf(prec uint) *big.Float { return new(big.Float).SetPrec(prec) }

func bigF(x float64, prec uint) *big.Float { return nf(prec).SetFloat64(x) }

func roundF(f *big.Float) float64 {
	r, _ := f.Float64()
	return r
}

var (
	constOnce sync.Once
	ln2Big    *big.Float
	ln10Big   *big.Float
	piBig     *big.Float
)

const constPrec = 1400

// atanhSeries is atanh(z) for small z.
func atanhSeries(z *big.Float, prec uint) *big.Float {
	z2 := nf(prec).Mul(z, z)
	term := nf(prec).Set(z)
	sum := nf(prec).Set(z)
	eps := -int(prec) - 8
	for n := int64(3); ; n += 2 {
		term.Mul(term, z2)
		t := nf(prec).Quo(term, nf(prec).SetInt64(n))
		if t.Sign() == 0 || t.MantExp(nil)-sum.MantExp(nil) < eps {
			break
		}
		sum.Add(sum, t)
	}
	return sum
}

// atanInv is atan(1/n) for an integer n > 1.
func atanInv(n int64, prec uint) *big.Float {
	x := nf(prec).Quo(nf(prec).SetInt64(1), nf(prec).SetInt64(n))
	x2 := nf(prec).Mul(x, x)
	term := nf(prec).Set(x)
	sum := nf(prec).Set(x)
	eps := -int(prec) - 8
	for k := int64(3); ; k += 2 {
		term.Mul(term, x2)
		t := nf(prec).Quo(term, nf(prec).SetInt64(k))
		if t.Sign() == 0 || t.MantExp(nil)-sum.MantExp(nil) < eps {
			break
		}
		if (k/2)%2 == 1 {
			sum.Sub(sum, t)
		} else {
			sum.Add(sum, t)
		}
	}
	return sum
}

func initConsts() {
	constOnce.Do(func() {
		p := uint(constPrec)
		ln2Big = nf(p).Mul(nf(p).SetInt64(2), atanhSeries(nf(p).Quo(nf(p).SetInt64(1), nf(p).SetInt64(3)), p))
		l125 := nf(p).Mul(nf(p).SetInt64(2), atanhSeries(nf(p).Quo(nf(p).SetInt64(1), nf(p).SetInt64(9)), p))
		ln10Big = nf(p).Add(nf(p).Mul(nf(p).SetInt64(3), ln2Big), l125)
		a := nf(p).Mul(nf(p).SetInt64(16), atanInv(5, p))
		b := nf(p).Mul(nf(p).SetInt64(4), atanInv(239, p))
		piBig = nf(p).Sub(a, b)
	})
}

// expBig is e**x.
func expBig(x *big.Float, prec uint) *big.Float {
	initConsts()
	p := prec + 24
	if x.Sign() == 0 {
		return nf(prec).SetInt64(1)
	}
	q := nf(p+40).Quo(x, ln2Big)
	qf, _ := q.Float64()
	k := int64(math.Round(qf))
	r := nf(p+40).Sub(x, nf(p+40).Mul(nf(p+40).SetInt64(k), ln2Big))
	r = nf(p).Set(r)
	r.SetMantExp(r, -8)
	sum := nf(p).SetInt64(1)
	term := nf(p).SetInt64(1)
	eps := -int(p) - 8
	for n := int64(1); n < 200; n++ {
		term.Mul(term, r)
		term.Quo(term, nf(p).SetInt64(n))
		if term.Sign() == 0 || term.MantExp(nil)-sum.MantExp(nil) < eps {
			break
		}
		sum.Add(sum, term)
	}
	for i := 0; i < 8; i++ {
		sum.Mul(sum, sum)
	}
	return nf(prec).SetMantExp(sum, int(k))
}

// logBig is ln x for x > 0.
func logBig(x *big.Float, prec uint) *big.Float {
	initConsts()
	p := prec + 24
	m := nf(p)
	e := x.MantExp(m) // x = m * 2**e, 0.5 <= m < 1
	if m.Cmp(nf(p).SetFloat64(0.7071067811865476)) < 0 {
		m.SetMantExp(m, 1)
		e--
	}
	one := nf(p).SetInt64(1)
	z := nf(p).Quo(nf(p).Sub(m, one), nf(p).Add(m, one))
	r := atanhSeries(z, p)
	r.SetMantExp(r, 1)
	if e != 0 {
		r.Add(r, nf(p).Mul(nf(p).SetInt64(int64(e)), ln2Big))
	}
	return nf(prec).Set(r)
}

// reduce returns r and the quadrant q such that x = q*pi/2 + r, |r| <= pi/4.
func reducePi2(x *big.Float, prec uint) (*big.Float, int) {
	initConsts()
	exp := x.MantExp(nil)
	if exp < 0 {
		exp = 0
	}
	p := prec + uint(exp) + 64
	halfPi := nf(constPrec).Set(piBig)
	halfPi.SetMantExp(halfPi, -1)
	q := nf(p).Quo(nf(p).Set(x), halfPi)
	qi, _ := q.Int(nil) // truncated
	// round to nearest
	frac := nf(p).Sub(q, nf(p).SetInt(qi))
	half := nf(p).SetFloat64(0.5)
	if frac.Cmp(half) > 0 {
		qi.Add(qi, big.NewInt(1))
	} else if frac.Cmp(nf(p).Neg(half)) < 0 {
		qi.Sub(qi, big.NewInt(1))
	}
	r := nf(p).Sub(nf(p).Set(x), nf(p).Mul(nf(p).SetInt(qi), halfPi))
	quad := int(new(big.Int).Mod(qi, big.NewInt(4)).Int64())
	return r, quad
}

func sinCosSeries(r *big.Float, prec uint) (s, c *big.Float) {
	p := prec + 16
	r2 := nf(p).Mul(r, r)
	s = nf(p).Set(r)
	c = nf(p).SetInt64(1)
	ts := nf(p).Set(r)
	tc := nf(p).SetInt64(1)
	eps := -int(p) - 8
	for n := int64(1); n < 200; n++ {
		tc.Mul(tc, r2)
		tc.Quo(tc, nf(p).SetInt64((2*n-1)*(2*n)))
		tc.Neg(tc)
		ts.Mul(ts, r2)
		ts.Quo(ts, nf(p).SetInt64((2*n)*(2*n+1)))
		ts.Neg(ts)
		if tc.Sign() != 0 {
			c.Add(c, tc)
		}
		if ts.Sign() != 0 {
			s.Add(s, ts)
		}
		if (tc.Sign() == 0 || tc.MantExp(nil) < -int(p)-8) && (ts.Sign() == 0 || ts.MantExp(nil)-s.MantExp(nil) < eps) {
			break
		}
	}
	return s, c
}

func sinCosBig(x *big.Float, prec uint) (sin, cos *big.Float) {
	r, q := reducePi2(x, prec)
	s, c := sinCosSeries(r, prec)
	switch q {
	case 0:
		return s, c
	case 1:
		return c, nf(prec).Neg(s)
	case 2:
		return nf(prec).Neg(s), nf(prec).Neg(c)
	}
	return nf(prec).Neg(c), s
}

// atanBig is atan x.
func atanBig(x *big.Float, prec uint) *big.Float {
	initConsts()
	p := prec + 24
	neg := x.Sign() < 0
	a := nf(p).Abs(x)
	one := nf(p).SetInt64(1)
	inv := false
	if a.Cmp(one) > 0 {
		a = nf(p).Quo(one, a)
		inv = true
	}
	for i := 0; i < 3; i++ {
		d := nf(p).Add(one, nf(p).Sqrt(nf(p).Add(one, nf(p).Mul(a, a))))
		a = nf(p).Quo(a, d)
	}
	a2 := nf(p).Mul(a, a)
	term := nf(p).Set(a)
	sum := nf(p).Set(a)
	eps := -int(p) - 8
	for n := int64(3); n < 400; n += 2 {
		term.Mul(term, a2)
		t := nf(p).Quo(term, nf(p).SetInt64(n))
		if t.Sign() == 0 || t.MantExp(nil)-sum.MantExp(nil) < eps {
			break
		}
		if (n/2)%2 == 1 {
			sum.Sub(sum, t)
		} else {
			sum.Add(sum, t)
		}
	}
	sum.SetMantExp(sum, 3)
	if inv {
		hp := nf(p).Set(piBig)
		hp.SetMantExp(hp, -1)
		sum = nf(p).Sub(hp, sum)
	}
	if neg {
		sum.Neg(sum)
	}
	return nf(prec).Set(sum)
}

func special(x float64) bool { return math.IsNaN(x) || math.IsInf(x, 0) || x == 0 }

// CrExp is exp(x).
func CrExp(x float64) float64 {
	if special(x) || x > 710 || x < -746 {
		return math.Exp(x)
	}
	return roundF(expBig(bigF(x, crPrec), crPrec))
}

// CrExp2 is 2**x.
func CrExp2(x float64) float64 {
	if special(x) || x == math.Trunc(x) || x > 1025 || x < -1080 {
		return math.Exp2(x)
	}
	initConsts()
	return roundF(expBig(nf(crPrec+16).Mul(bigF(x, crPrec+16), ln2Big), crPrec))
}

// CrExpm1 is exp(x) - 1.
func CrExpm1(x float64) float64 {
	if special(x) || x > 710 || x < -40 {
		return math.Expm1(x)
	}
	p := uint(crPrec + 8)
	if exp := bigF(x, p).MantExp(nil); exp < 0 {
		p += uint(-2 * exp)
	}
	e := expBig(bigF(x, p), p)
	return roundF(nf(p).Sub(e, nf(p).SetInt64(1)))
}

// CrLog is the natural logarithm of x > 0.
func CrLog(x float64) float64 {
	if special(x) || x < 0 || x == 1 {
		return math.Log(x)
	}
	return roundF(logBig(bigF(x, crPrec), crPrec))
}

// CrLog2 is the base 2 logarithm.
func CrLog2(x float64) float64 {
	if special(x) || x < 0 || x == 1 {
		return math.Log2(x)
	}
	initConsts()
	l := logBig(bigF(x, crPrec+16), crPrec+16)
	return roundF(nf(crPrec).Quo(l, ln2Big))
}

// CrLog10 is the base 10 logarithm.
func CrLog10(x float64) float64 {
	if special(x) || x < 0 || x == 1 {
		return math.Log10(x)
	}
	initConsts()
	l := logBig(bigF(x, crPrec+16), crPrec+16)
	return roundF(nf(crPrec).Quo(l, ln10Big))
}

// CrLog1p is log(1 + x).
func CrLog1p(x float64) float64 {
	if special(x) || x <= -1 {
		return math.Log1p(x)
	}
	p := uint(crPrec + 16)
	if exp := bigF(x, p).MantExp(nil); exp < 0 {
		p += uint(-exp) + 8
	}
	one := nf(p).Add(nf(p).SetInt64(1), bigF(x, p))
	return roundF(logBig(one, p))
}

// CrPow is x**y for finite x and y.
func CrPow(x, y float64) float64 {
	if special(x) || special(y) || y == 1 || x == 1 {
		return math.Pow(x, y)
	}
	neg := false
	if x < 0 {
		if y != math.Trunc(y) {
			return math.Pow(x, y)
		}
		if math.Mod(y, 2) != 0 {
			neg = true
		}
		x = -x
	}
	var r float64
	if y == math.Trunc(y) && math.Abs(y) <= 64 {
		// an exact power of a float
		n := int(math.Abs(y))
		p := uint(54*64 + 64)
		acc := nf(p).SetInt64(1)
		b := bigF(x, p)
		for i := 0; i < n; i++ {
			acc.Mul(acc, b)
		}
		if y < 0 {
			acc = nf(p).Quo(nf(p).SetInt64(1), acc)
		}
		r = roundF(acc)
	} else {
		lx := math.Log(x) * y
		if lx > 745 || lx < -750 {
			return signed(math.Pow(x, y), neg)
		}
		p := uint(crPrec + 24)
		l := logBig(bigF(x, p), p)
		l.Mul(l, bigF(y, p))
		r = roundF(expBig(l, p))
	}
	return signed(r, neg)
}

func signed(r float64, neg bool) float64 {
	if neg {
		return -r
	}
	return r
}

// CrSin, CrCos and CrTan are the trigonometric functions of a finite x.
func CrSin(x float64) float64 {
	if special(x) {
		return math.Sin(x)
	}
	if math.Abs(x) < 1e-8 {
		// x - x**3/6 rounds to x for these
		return roundF(sinSmall(x))
	}
	s, _ := sinCosBig(bigF(x, crPrec), crPrec)
	return roundF(s)
}

func sinSmall(x float64) *big.Float {
	s, _ := sinCosSeries(bigF(x, crPrec), crPrec)
	return s
}

func CrCos(x float64) float64 {
	if special(x) {
		return math.Cos(x)
	}
	_, c := sinCosBig(bigF(x, crPrec), crPrec)
	return roundF(c)
}

func CrTan(x float64) float64 {
	if special(x) {
		return math.Tan(x)
	}
	s, c := sinCosBig(bigF(x, crPrec), crPrec)
	return roundF(nf(crPrec).Quo(s, c))
}

// CrAtan is atan(x).
func CrAtan(x float64) float64 {
	if special(x) {
		return math.Atan(x)
	}
	return roundF(atanBig(bigF(x, crPrec), crPrec))
}

// CrAsin is asin(x) for |x| <= 1.
func CrAsin(x float64) float64 {
	if special(x) || math.Abs(x) > 1 {
		return math.Asin(x)
	}
	if math.Abs(x) == 1 {
		return math.Asin(x)
	}
	p := uint(crPrec + 24)
	b := bigF(x, p)
	one := nf(p).SetInt64(1)
	d := nf(p).Sqrt(nf(p).Mul(nf(p).Sub(one, b), nf(p).Add(one, b)))
	return roundF(atanBig(nf(p).Quo(b, d), crPrec))
}

// CrAcos is acos(x) for |x| <= 1.
func CrAcos(x float64) float64 {
	if special(x) || math.Abs(x) >= 1 {
		return math.Acos(x)
	}
	initConsts()
	p := uint(crPrec + 24)
	b := bigF(x, p)
	one := nf(p).SetInt64(1)
	d := nf(p).Sqrt(nf(p).Mul(nf(p).Sub(one, b), nf(p).Add(one, b)))
	a := atanBig(nf(p).Quo(d, b), p) // atan(sqrt(1-x^2)/x), in (-pi/2, pi/2)
	if x < 0 {
		a = nf(p).Add(a, piBig)
	}
	return roundF(a)
}

// CrAtan2 is atan2(y, x) for finite nonzero arguments.
func CrAtan2(y, x float64) float64 {
	if special(x) || special(y) {
		return math.Atan2(y, x)
	}
	initConsts()
	p := uint(crPrec + 24)
	a := atanBig(nf(p).Quo(bigF(y, p), bigF(x, p)), p)
	if x < 0 {
		if y > 0 {
			a = nf(p).Add(a, piBig)
		} else {
			a = nf(p).Sub(a, piBig)
		}
	}
	return roundF(a)
}

func hyperbolic(x float64, which int) float64 {
	p := uint(crPrec + 16)
	exp := bigF(x, p).MantExp(nil)
	if exp < 0 {
		p += uint(-3 * exp)
	}
	b := bigF(x, p)
	e := expBig(b, p)
	inv := nf(p).Quo(nf(p).SetInt64(1), e)
	switch which {
	case 0: // sinh
		return roundF(nf(p).Quo(nf(p).Sub(e, inv), nf(p).SetInt64(2)))
	case 1: // cosh
		return roundF(nf(p).Quo(nf(p).Add(e, inv), nf(p).SetInt64(2)))
	}
	return roundF(nf(p).Quo(nf(p).Sub(e, inv), nf(p).Add(e, inv)))
}

func CrSinh(x float64) float64 {
	if special(x) || math.Abs(x) > 710 {
		return math.Sinh(x)
	}
	return hyperbolic(x, 0)
}

func CrCosh(x float64) float64 {
	if special(x) || math.Abs(x) > 710 {
		return math.Cosh(x)
	}
	return hyperbolic(x, 1)
}

func CrTanh(x float64) float64 {
	if special(x) || math.Abs(x) > 20 {
		return math.Tanh(x)
	}
	return hyperbolic(x, 2)
}

// CrAsinh is asinh(x).
func CrAsinh(x float64) float64 {
	if special(x) || math.Abs(x) > 1e150 {
		return math.Asinh(x)
	}
	p := uint(crPrec + 16)
	if exp := bigF(x, p).MantExp(nil); exp < 0 {
		p += uint(-3 * exp)
	}
	a := nf(p).Abs(bigF(x, p))
	s := nf(p).Sqrt(nf(p).Add(nf(p).Mul(a, a), nf(p).SetInt64(1)))
	r := roundF(logBig(nf(p).Add(a, s), p))
	return math.Copysign(r, x)
}

// CrAcosh is acosh(x) for x > 1.
func CrAcosh(x float64) float64 {
	if special(x) || x <= 1 || x > 1e150 {
		return math.Acosh(x)
	}
	p := uint(crPrec + 24)
	b := bigF(x, p)
	one := nf(p).SetInt64(1)
	s := nf(p).Sqrt(nf(p).Mul(nf(p).Sub(b, one), nf(p).Add(b, one)))
	return roundF(logBig(nf(p).Add(b, s), p))
}

// CrAtanh is atanh(x) for |x| < 1.
func CrAtanh(x float64) float64 {
	if special(x) || math.Abs(x) >= 1 {
		return math.Atanh(x)
	}
	p := uint(crPrec + 16)
	if exp := bigF(x, p).MantExp(nil); exp < 0 {
		p += uint(-3 * exp)
	}
	b := bigF(x, p)
	one := nf(p).SetInt64(1)
	q := nf(p).Quo(nf(p).Add(one, b), nf(p).Sub(one, b))
	l := logBig(q, p)
	return roundF(nf(p).Quo(l, nf(p).SetInt64(2)))
}

// CrCbrt is the cube root.
func CrCbrt(x float64) float64 {
	if special(x) {
		return math.Cbrt(x)
	}
	p := uint(crPrec + 24)
	a := nf(p).Abs(bigF(x, p))
	// Newton from the float estimate
	y := nf(p).SetFloat64(math.Cbrt(math.Abs(x)))
	three := nf(p).SetInt64(3)
	for i := 0; i < 5; i++ {
		y2 := nf(p).Mul(y, y)
		y = nf(p).Quo(nf(p).Add(nf(p).Mul(nf(p).SetInt64(2), y), nf(p).Quo(a, y2)), three)
	}
	return math.Copysign(roundF(y), x)
}

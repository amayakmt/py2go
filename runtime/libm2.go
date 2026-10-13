package runtime

import (
	"math"
	"math/big"
)

// The functions glibc and CPython compute with their own algorithms, which
// are not correctly rounded and so are ported, not computed exactly.

// glibcLog10 is __ieee754_log10 of glibc.
func glibcLog10(x float64) float64 {
	const (
		two54     = 1.80143985094819840000e+16
		ivln10    = 4.34294481903251816668e-01
		log10_2hi = 3.01029995663611771306e-01
		log10_2lo = 3.69423907715893078616e-13
	)
	if math.IsNaN(x) || math.IsInf(x, 0) || x <= 0 {
		return math.Log10(x)
	}
	bits := math.Float64bits(x)
	hx := int32(bits >> 32)
	k := int32(0)
	if hx < 0x00100000 {
		k -= 54
		x *= two54
		bits = math.Float64bits(x)
		hx = int32(bits >> 32)
	}
	k += (hx >> 20) - 1023
	i := int32((uint32(k) & 0x80000000) >> 31)
	hx = (hx & 0x000fffff) | ((0x3ff - i) << 20)
	y := float64(k + i)
	x = math.Float64frombits(uint64(uint32(hx))<<32 | bits&0xffffffff)
	z := y*log10_2lo + ivln10*CrLog(x)
	return z + y*log10_2hi
}

// glibcCbrt is s_cbrt.c of glibc.
func glibcCbrt(x float64) float64 {
	const (
		cbrt2    = 1.2599210498948731648
		sqrCbrt2 = 1.5874010519681994748
	)
	factor := [5]float64{1.0 / sqrCbrt2, 1.0 / cbrt2, 1.0, cbrt2, sqrCbrt2}
	if x == 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return x + x
	}
	xm, xe := math.Frexp(math.Abs(x))
	u := 0.354895765043919860 + ((1.50819193781584896 + ((-2.11499494167371287 + ((2.44693122563534430 + ((-1.83469277483613086 + (0.784932344976639262-0.145263899385486377*xm)*xm) * xm)) * xm)) * xm)) * xm)
	t2 := u * u * u
	ym := u * (t2 + 2.0*xm) / (2.0*t2 + xm) * factor[2+xe%3]
	if x < 0 {
		ym = -ym
	}
	return math.Ldexp(ym, xe/3)
}

// fdlibm's sinh, cosh and tanh on top of expm1 and exp.
func glibcSinh(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	h := 0.5
	if x < 0 {
		h = -0.5
	}
	ax := math.Abs(x)
	if ax < 22 {
		if ax < 3.725290298461914e-09 {
			return x
		}
		t := math.Expm1(ax)
		if ax < 1 {
			return h * (2.0*t - t*t/(t+1.0))
		}
		return h * (t + t/(t+1.0))
	}
	if ax < 7.10475860073943863426e+02 {
		return h * CrExp(ax)
	}
	if ax <= 7.10758604650778879738e+02 {
		w := CrExp(0.5 * ax)
		t := h * w
		return t * w
	}
	return x * 1.0e307 * 1.0e307
}

func glibcCosh(x float64) float64 {
	if math.IsNaN(x) {
		return x
	}
	if math.IsInf(x, 0) {
		return math.Inf(1)
	}
	ax := math.Abs(x)
	if ax < 0.34657359027997264 {
		t := math.Expm1(ax)
		w := 1.0 + t
		if ax < 2.7755575615628914e-17 {
			return w
		}
		return 1.0 + (t*t)/(w+w)
	}
	if ax < 22 {
		t := CrExp(ax)
		return 0.5*t + 0.5/t
	}
	if ax < 7.10475860073943863426e+02 {
		return 0.5 * CrExp(ax)
	}
	if ax <= 7.10758604650778879738e+02 {
		w := CrExp(0.5 * ax)
		t := 0.5 * w
		return t * w
	}
	return math.Inf(1)
}

func glibcTanh(x float64) float64 {
	if math.IsNaN(x) {
		return x
	}
	if math.IsInf(x, 0) {
		return math.Copysign(1, x)
	}
	ax := math.Abs(x)
	var z float64
	if ax < 22 {
		if ax < 2.7755575615628914e-17 {
			return x * (1.0 + x)
		}
		if ax >= 1 {
			t := math.Expm1(2 * ax)
			z = 1.0 - 2.0/(t+2.0)
		} else {
			t := math.Expm1(-2 * ax)
			z = -t / (t + 2.0)
		}
	} else {
		z = 1.0
	}
	if x >= 0 {
		return z
	}
	return -z
}

// ---- CPython's own hypot, gamma and lgamma ----

func dlMul(x, y float64) (hi, lo float64) {
	hi = x * y
	lo = math.FMA(x, y, -hi)
	return
}

// vectorNorm is vector_norm of mathmodule.c.
func vectorNorm(vec []float64, max float64) float64 {
	n := len(vec)
	if max == 0 || n <= 1 {
		return max
	}
	_, maxE := math.Frexp(max)
	if maxE < -1023 {
		// subnormal: scale up, the result is scaled back
		sc := make([]float64, n)
		for i, v := range vec {
			sc[i] = v / 2.2250738585072014e-308 / 2.220446049250313e-16
		}
		return vectorNorm(sc, max/2.2250738585072014e-308/2.220446049250313e-16) * 2.2250738585072014e-308 * 2.220446049250313e-16
	}
	scale := math.Ldexp(1.0, -maxE)
	csum := 1.0
	frac1, frac2 := 0.0, 0.0
	for _, v := range vec {
		x := v * scale
		hi, lo := dlMul(x, x)
		sm := csum + hi
		frac1 += (csum - sm) + hi
		csum = sm
		frac2 += lo
	}
	h := math.Sqrt(csum - 1.0 + (frac1 + frac2))
	hi, lo := dlMul(-h, h)
	sm := csum + hi
	frac1 += (csum - sm) + hi
	csum = sm
	frac2 += lo
	x := csum - 1.0 + (frac1 + frac2)
	h += x / (2.0 * h)
	return h / scale
}

const (
	lanczosN           = 13
	lanczosG           = 6.024680040776729583740234375
	lanczosGMinusHalf  = 5.524680040776729583740234375
	logPi              = 1.144729885849400174143427351353058711647294812915311571513623071472137769884826079783623270275489708
	lanczosNumCoeffs0  = 23531376880.410759688572007674451636754734846804940
	lanczosNumCoeffs1  = 42919803642.649098768957899047001988850926355848959
	lanczosNumCoeffs2  = 35711959237.355668049440185451547166705960488635843
	lanczosNumCoeffs3  = 17921034426.037209699919755754458931112671403265390
	lanczosNumCoeffs4  = 6039542586.3520280050642916443072979210699388420708
	lanczosNumCoeffs5  = 1439720407.3117216736632230727949123939715485786772
	lanczosNumCoeffs6  = 248874557.86205415651146038641322942321632125127801
	lanczosNumCoeffs7  = 31426415.585400194380614231628318205362874684987640
	lanczosNumCoeffs8  = 2876370.6289353724412254090516208496135991145378768
	lanczosNumCoeffs9  = 186056.26539522349504029498971604569928220784236328
	lanczosNumCoeffs10 = 8071.6720023658162106380029022722506138218516325024
	lanczosNumCoeffs11 = 210.82427775157934587250973392071336271166969580291
	lanczosNumCoeffs12 = 2.5066282746310002701649081771338373386264310793408
)

var lanczosNum = [lanczosN]float64{lanczosNumCoeffs0, lanczosNumCoeffs1, lanczosNumCoeffs2, lanczosNumCoeffs3, lanczosNumCoeffs4, lanczosNumCoeffs5, lanczosNumCoeffs6, lanczosNumCoeffs7, lanczosNumCoeffs8, lanczosNumCoeffs9, lanczosNumCoeffs10, lanczosNumCoeffs11, lanczosNumCoeffs12}
var lanczosDen = [lanczosN]float64{0.0, 39916800.0, 120543840.0, 150917976.0, 105258076.0, 45995730.0, 13339535.0, 2637558.0, 357423.0, 32670.0, 1925.0, 66.0, 1.0}

func lanczosSum(x float64) float64 {
	num, den := 0.0, 0.0
	if x < 5.0 {
		for i := lanczosN - 1; i >= 0; i-- {
			num = num*x + lanczosNum[i]
			den = den*x + lanczosDen[i]
		}
	} else {
		for i := 0; i < lanczosN; i++ {
			num = num/x + lanczosNum[i]
			den = den/x + lanczosDen[i]
		}
	}
	return num / den
}

func sinPi(x float64) float64 {
	y := math.Mod(math.Abs(x), 2.0)
	n := int(math.Round(2.0 * y))
	var r float64
	switch n {
	case 0:
		r = CrSin(math.Pi * y)
	case 1:
		r = CrCos(math.Pi * (y - 0.5))
	case 2:
		r = CrSin(math.Pi * (1.0 - y))
	case 3:
		r = -CrCos(math.Pi * (y - 1.5))
	case 4:
		r = CrSin(math.Pi * (y - 2.0))
	}
	return math.Copysign(1.0, x) * r
}

var gammaIntegral = [23]float64{1.0, 1.0, 2.0, 6.0, 24.0, 120.0, 720.0, 5040.0, 40320.0, 362880.0, 3628800.0, 39916800.0, 479001600.0, 6227020800.0, 87178291200.0, 1307674368000.0, 20922789888000.0, 355687428096000.0, 6402373705728000.0, 121645100408832000.0, 2432902008176640000.0, 51090942171709440000.0, 1124000727777607680000.0}

// pyGamma is m_tgamma of mathmodule.c; the errors are the caller's.
func pyGamma(x float64) (float64, error2) {
	if math.IsNaN(x) {
		return x, 0
	}
	if math.IsInf(x, 0) {
		if x > 0 {
			return x, 0
		}
		return 0, errDomain
	}
	if x == 0 {
		return 0, errDomain
	}
	if x == math.Floor(x) {
		if x < 0 {
			return 0, errDomain
		}
		if x <= 23 {
			return gammaIntegral[int(x)-1], 0
		}
	}
	absx := math.Abs(x)
	if absx < 1e-20 {
		r := 1.0 / x
		if math.IsInf(r, 0) {
			return 0, errRange
		}
		return r, 0
	}
	if absx > 200 {
		if x < 0 {
			return 0.0 / sinPi(x), 0
		}
		return 0, errRange
	}
	y := absx + lanczosGMinusHalf
	var q, z float64
	if absx > lanczosGMinusHalf {
		q = y - absx
		z = q - lanczosGMinusHalf
	} else {
		q = y - lanczosGMinusHalf
		z = q - absx
	}
	z = z * lanczosG / y
	var r float64
	if x < 0 {
		r = -math.Pi / sinPi(absx) / absx * CrExp(y) / lanczosSum(absx)
		r -= z * r
		if absx < 140.0 {
			r /= CrPow(y, absx-0.5)
		} else {
			sp := CrPow(y, absx/2.0-0.25)
			r /= sp
			r /= sp
		}
	} else {
		r = lanczosSum(absx) / CrExp(y)
		r += z * r
		if absx < 140.0 {
			r *= CrPow(y, absx-0.5)
		} else {
			sp := CrPow(y, absx/2.0-0.25)
			r *= sp
			r *= sp
		}
	}
	if math.IsInf(r, 0) {
		return 0, errRange
	}
	return r, 0
}

type error2 int

const (
	errDomain error2 = 1
	errRange  error2 = 2
)

// pyLgamma is m_lgamma of mathmodule.c.
func pyLgamma(x float64) (float64, error2) {
	if math.IsNaN(x) {
		return x, 0
	}
	if math.IsInf(x, 0) {
		return math.Inf(1), 0
	}
	if x == math.Floor(x) && x <= 2.0 {
		if x <= 0.0 {
			return 0, errDomain
		}
		return 0.0, 0
	}
	absx := math.Abs(x)
	if absx < 1e-20 {
		return -CrLog(absx), 0
	}
	r := CrLog(lanczosSum(absx)) - lanczosG
	r += (absx - 0.5) * (CrLog(absx+lanczosG-0.5) - 1)
	if x < 0.0 {
		r = logPi - r - CrLog(math.Abs(sinPi(absx))) - CrLog(absx)
	}
	if math.IsInf(r, 0) {
		return 0, errRange
	}
	return r, 0
}

var _ = big.NewInt

// ---- erf and erfc as CPython computes them ----

const sqrtPi = 1.772453850905516027298167483341145182798

func erfSeries(x float64) float64 {
	x2 := x * x
	acc := 0.0
	fk := 25.0 + 0.5
	for i := 0; i < 25; i++ {
		acc = 2.0 + x2*acc/fk
		fk -= 1.0
	}
	return acc * x * CrExp(-x2) / sqrtPi
}

func erfcContFrac(x float64) float64 {
	if x >= 30.0 {
		return 0.0
	}
	x2 := x * x
	a, da := 0.0, 0.5
	p, pLast := 1.0, 0.0
	q, qLast := da+x2, 1.0
	for i := 0; i < 50; i++ {
		a += da
		da += 2.0
		b := da + x2
		temp := p
		p = b*p - a*pLast
		pLast = temp
		temp = q
		q = b*q - a*qLast
		qLast = temp
	}
	return p / q * x * CrExp(-x2) / sqrtPi
}

func pyErf(x float64) float64 {
	if math.IsNaN(x) {
		return x
	}
	if math.IsInf(x, 0) {
		return math.Copysign(1, x)
	}
	if math.Abs(x) < 1.5 {
		return erfSeries(x)
	}
	cf := erfcContFrac(math.Abs(x))
	if x > 0 {
		return 1.0 - cf
	}
	return cf - 1.0
}

func pyErfc(x float64) float64 {
	if math.IsNaN(x) {
		return x
	}
	if math.IsInf(x, 0) {
		if x > 0 {
			return 0
		}
		return 2
	}
	ax := math.Abs(x)
	if ax < 1.5 {
		return 1.0 - erfSeries(x)
	}
	cf := erfcContFrac(ax)
	if x > 0 {
		return cf
	}
	return 2.0 - cf
}

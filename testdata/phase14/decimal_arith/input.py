from decimal import *
import decimal
import math

D = Decimal
print(D("1.5"), D(" 12 "), D("-0"), D("1e5"), D("1E-7"), D("0.000001"), D("0.0000001"), D(".5"), D("5."), D("+inf"), D("-Infinity"), D("nan"), D("-sNaN123"), D(0), D(-5), D(10 ** 30), D(True))
print(repr(D("3.14")), repr(D(7)), D((1, (3, 1, 4), -2)), D("1_000.5"), D(D("2.50")), D([0, [1, 2], 3]).as_tuple(), D("123.4500").as_tuple())
print(D(0.1), D(0.5), D(1e22), D(-2.5e-10), Decimal.from_float(1 / 3), D(float("inf")), D(float("nan")).is_nan())
print(D("1.1") + D("2.2"), D("1.1") - D("2.20"), D("1.10") * D("2.2"), D("10") / D("4"), D("1") / D("3"), D("2") / D("3"), D("7") // D("2"), D("-7") // D("2"), D("7") % D("-2"), D("-7") % D("2"), divmod(D("7.5"), D("2")))
print(D("2") ** 10, D("2") ** -2, D("2.5") ** 3, D("2") ** D("0.5"), D("10") ** D("-3"), D("0") ** 0 if False else "", D("1.5") ** D("2.5"), abs(D("-3.5")), -D("1.5"), +D("1.50"), D("-0") + 0, 0 - D("0"))
print(1 + D("2"), 2 * D("1.5"), 3 - D("0.5"), 1 / D("8"), 7 // D("2"), 7 % D("4"), 2 ** D("3"), D("1.5") + 1, D("1.5") * 2, D("5") / 2, D("5") // 2)
print(D("1.23456789") + D("0"), +D("1.23456789"), D("123456789012345678901234567890.5") + 1, D("1e30") + 1, D("1e-30") + 1)
print(D("1.0") == D("1.00"), D("1") == 1, D("1.5") == 1.5, D("1.5") < 2, D("2") > D("1.9"), D("nan") == D("nan"), D("1") != D("2"), D("0") == D("-0"), D("1.5") == "x", D("2") <= 2, D("1") >= 3)
print(hash(D("1")) == hash(1), hash(D("1.5")) == hash(1.5), hash(D("0.1")) == hash(0.1), {D("1.0"): "a"}[D("1")], sorted([D("3"), D("1.5"), D("2")]), max(D("1"), D("2.5")), min(D("1"), D("2.5")), sum([D("0.1")] * 10))
print(int(D("3.9")), int(D("-3.9")), float(D("2.5")), round(D("2.5")), round(D("3.5")), round(D("2.567"), 2), round(D("1234"), -2), math.floor(D("2.5")) if False else "", bool(D("0")), bool(D("0.0001")), D("12.5").as_integer_ratio(), D("3").is_signed(), D("-3").is_signed())
print(D("2.50").quantize(D("0.1")), D("2.55").quantize(D("0.1")), D("2.65").quantize(D("0.1")), D("1234.5").quantize(D("1e2")), D("2.5").quantize(D("0.001")), D("-1.5").quantize(D(1), rounding=ROUND_FLOOR), D("1.5").quantize(D(1), rounding=ROUND_CEILING), D("1.45").quantize(D("0.1"), ROUND_HALF_DOWN), D("1.55").quantize(D("0.1"), ROUND_UP), D("1.55").quantize(D("0.1"), ROUND_DOWN), D("1.05").quantize(D("0.1"), ROUND_05UP))
print(D("2.5").to_integral_value(), D("3.5").to_integral(), D("-2.5").to_integral_value(ROUND_HALF_UP), D("123.456").normalize(), D("1.2000").normalize(), D("100").normalize(), D("0.000").normalize(), D("1E+2").to_integral_exact())
print(D("2").sqrt(), D("16").sqrt(), D("0.25").sqrt(), D("2").exp(), D("1").exp(), D("10").ln(), D("100").log10(), D("2").log10(), D("0.5").ln(), D("12.345").logb(), D("1000").scaleb(2), D("-0").sqrt())
print(D("7").fma(D("3"), D("1")), D("5").copy_sign(D("-1")), D("-5").copy_abs(), D("5").copy_negate(), D("1.5").compare(D("2")), D("2").compare_total(D("2.0")), D("1").max(D("1.0")), D("1.0").max(D("1")), D("3").min_mag(D("-2")), D("5").next_plus(), D("5").next_minus(), D("1").next_toward(D("2")))
print(D("12.30").adjusted(), D("0.00123").adjusted(), D("Inf").is_infinite(), D("5").is_finite(), D("nan").is_nan(), D("snan").is_snan(), D("0e5").is_zero(), D("1e-999").is_subnormal(), D("1").is_normal(), D("1.5").number_class(), D("-0").number_class(), D("inf").number_class(), D("1").radix(), D("100").same_quantum(D("1")))
print(D("5") & D("1") if False else "", D("101").logical_and(D("110")), D("101").logical_or(D("110")), D("101").logical_xor(D("110")), D("101").logical_invert(Context(prec=5)), D("12345").rotate(2), D("12345").shift(-2), D("12345").shift(2, Context(prec=6)) if False else "")
for spec in ["", "f", ".2f", "10.3f", "<10.1f", "^10.1f", ">+10.2f", "010.2f", ",.2f", "_.1f", "e", ".3e", "E", "g", ".3g", "%", ".1%", "G", "+", " f", ".0f", ",", "010,.1f", "*^12.2f", ".10f"]:
    print(repr(spec), format(D("1234567.891"), spec), format(D("-0.000123456"), spec), format(D("0"), spec), format(D("1e10"), spec))
print(format(D("inf"), "f"), format(D("-nan"), ".2f"), format(D("12345678901234567890.123"), ".2f"), "{:>8.2f}".format(D("3.14159")), f"{D('2.5'):.3e}", "%s %r" % (D("1.5"), D("1.5")), str(D("1E+3")), str(D("1E-8")), str(D("12E+4")), D("1E+3").to_eng_string(), D("12E+4").to_eng_string(), D("0.00123E-5").to_eng_string())
ctx = getcontext()
print(ctx.prec, ctx.rounding, ctx.Emin, ctx.Emax, ctx.capitals, ctx.clamp, ctx.traps[InvalidOperation], ctx.traps[Inexact], ctx.flags[Inexact], ctx)
with localcontext() as c:
    c.prec = 5
    print(D(1) / D(3), D("123456789") + 0, c.prec, getcontext().prec, c.flags[Inexact], c.flags[Rounded])
    c.rounding = ROUND_DOWN
    print(D(2) / D(3), D("1.99999999") * 1)
    c.prec = 3
    print(D(2).sqrt(), D(10).ln(), D(1).exp(), D(3) ** D("0.5"), D(2) ** 10, D("1234") + D("1"), c.flags[Inexact], c.flags[Rounded])
print(getcontext().prec)
with localcontext(Context(prec=4, rounding=ROUND_HALF_UP)) as c:
    print(D("1.23456"), +D("1.23456"), D("2.5").quantize(D(1)), c.create_decimal("1.23456"), c.create_decimal_from_float(0.1), c.plus(D("9.9999")), c.divide(1, 3), c.add(1, 2), c.multiply(D("1.5"), 3), c.power(2, 10), c.sqrt(2), c.minus(1), c.abs(-3), c.compare(1, 2), c.to_sci_string(D("1E+5")))
with localcontext() as c:
    c.prec = 2
    c.Emax = 99
    c.Emin = -99
    print(D("9.9e99") * 10 if False else "", D("1e100") + 0 if False else "")
    c.traps[Overflow] = 0
    c.traps[Underflow] = 0
    print(D("9e99") * 10, D("1e-99") / 1000, D("1e-101"), D("5e-100") * 1, [s.__name__ for s, v in c.flags.items() if v])
    c.clear_flags()
    print(D("1") / 7, c.flags[Inexact], c.flags[Rounded], c.flags[Clamped])
cases = [
    ('D(1) / 0', lambda: D(1) / 0),
    ('D(0) / 0', lambda: D(0) / 0),
    ("D('abc')", lambda: D('abc')),
    ("D('1e999999999999') * D('1e999999999999')", lambda: D('1e999999999999') * D('1e999999999999')),
    ("D('nan').quantize(D(1)) + D('snan')", lambda: D('nan').quantize(D(1)) + D('snan')),
    ('D(1) % 0', lambda: D(1) % 0),
    ('D(-1).sqrt()', lambda: D(-1).sqrt()),
    ("D(2) ** D('1.5') * D('1e9999999') ** 2", lambda: D(2) ** D('1.5') * D('1e9999999') ** 2),
    ("D('1.5').quantize(D('1e50'))", lambda: D('1.5').quantize(D('1e50'))),
    ('D(1) // 0', lambda: D(1) // 0),
    ("D('inf') - D('inf')", lambda: D('inf') - D('inf')),
    ('D(0) ** 0', lambda: D(0) ** 0),
    ("D(-8) ** D('0.5')", lambda: D(-8) ** D('0.5')),
    ('D(1).ln() + D(-1).ln()', lambda: D(1).ln() + D(-1).ln()),
    ('divmod(D(1), 0)', lambda: divmod(D(1), 0)),
    ('D(1) + 1.5', lambda: D(1) + 1.5),
    ('D(1) < 1.5', lambda: D(1) < 1.5),
    ('D(1.5) == 1.5', lambda: D(1.5) == 1.5),
    ("int(D('nan'))", lambda: int(D('nan'))),
    ("int(D('inf'))", lambda: int(D('inf'))),
    ("float(D('snan'))", lambda: float(D('snan'))),
    ('D(1, 2)', lambda: D(1, 2)),
    ('D((2, (1,), 0))', lambda: D((2, (1,), 0))),
    ('D((0, (1, 10), 0))', lambda: D((0, (1, 10), 0))),
    ("D('1').quantize('x')", lambda: D('1').quantize('x')),
    ("D('1').compare(None)", lambda: D('1').compare(None)),
    ("hash(D('snan'))", lambda: hash(D('snan'))),
    ("round(D('inf'))", lambda: round(D('inf'))),
    ('D(1) + None', lambda: D(1) + None),
    ('D(object())', lambda: D(object())),
]
for label, fn in cases:
    try:
        print(label, "->", fn())
    except Exception as e:
        print(label, "->", type(e).__name__, str(e).replace("'decimal.Decimal'", "'Decimal'"))
c = Context(prec=9, rounding=ROUND_HALF_EVEN, traps=[InvalidOperation])
c.traps[DivisionByZero] = 0
print(c.divide(1, 0), c.divide(0, 0) if False else "", c.flags[DivisionByZero], c.copy().prec, repr(c).startswith("Context(prec=9"), BasicContext.prec, ExtendedContext.rounding, DefaultContext.traps[Overflow])
setcontext(ExtendedContext)
print(D(1) / D(0), D(0) / D(0), D(1) / 3, getcontext().prec)
setcontext(DefaultContext)
print(getcontext().prec, issubclass(DivisionByZero, ZeroDivisionError), issubclass(FloatOperation, TypeError), issubclass(Overflow, Inexact), issubclass(InvalidOperation, ArithmeticError), MAX_PREC > 10 ** 17, decimal.MIN_EMIN)

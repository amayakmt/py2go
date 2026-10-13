import random
from decimal import *

random.seed(4)
modes = [ROUND_DOWN, ROUND_HALF_UP, ROUND_HALF_EVEN, ROUND_CEILING, ROUND_FLOOR, ROUND_UP, ROUND_HALF_DOWN, ROUND_05UP]

def rnd():
    k = random.randrange(8)
    if k == 0:
        return Decimal(random.randrange(-1000, 1000))
    if k == 1:
        return Decimal(random.randrange(-10**12, 10**12)) / Decimal(10 ** random.randrange(0, 9))
    if k == 2:
        s = "".join(random.choice("0123456789") for _ in range(random.randrange(1, 30)))
        return Decimal(s + "e" + str(random.randrange(-30, 30)))
    if k == 3:
        return Decimal(random.choice(["0", "-0", "0.00", "1", "-1", "0.5", "10", "1e5", "1e-5", "Infinity", "-Infinity", "NaN"]))
    if k == 4:
        return Decimal(random.random() * 10 ** random.randrange(-5, 12))
    return Decimal("%d.%d" % (random.randrange(0, 99999), random.randrange(0, 99999)))

ops = ["add", "sub", "mul", "div", "floordiv", "mod", "pow", "quantize", "compare", "max", "min", "rem_near", "copy_sign", "fma"]
out = []
for i in range(700):
    ctx = Context(prec=random.choice([1, 2, 3, 5, 9, 15, 28, 40]), rounding=random.choice(modes), Emin=random.choice([-999999, -50, -5]), Emax=random.choice([999999, 50, 5]), traps=[], clamp=random.choice([0, 0, 1]))
    a, b, c = rnd(), rnd(), rnd()
    op = random.choice(ops)
    try:
        if op == "add": r = ctx.add(a, b)
        elif op == "sub": r = ctx.subtract(a, b)
        elif op == "mul": r = ctx.multiply(a, b)
        elif op == "div": r = ctx.divide(a, b)
        elif op == "floordiv": r = ctx.divide_int(a, b)
        elif op == "mod": r = ctx.remainder(a, b)
        elif op == "pow":
            if abs(b) > 40 or b != b: b = Decimal(random.randrange(-5, 6))
            r = ctx.power(a, b)
        elif op == "quantize": r = ctx.quantize(a, Decimal(1).scaleb(random.randrange(-6, 6)))
        elif op == "compare": r = ctx.compare(a, b)
        elif op == "max": r = ctx.max(a, b)
        elif op == "min": r = ctx.min(a, b)
        elif op == "rem_near": r = ctx.remainder_near(a, b)
        elif op == "copy_sign": r = ctx.copy_sign(a, b)
        else: r = ctx.fma(a, b, c)
        fl = "".join(sorted(f.__name__[:3] for f, v in ctx.flags.items() if v))
        out.append("%d %s %s %s [p=%d %s Emin=%d Emax=%d cl=%d] -> %s %s" % (i, op, a, b, ctx.prec, ctx.rounding[6:], ctx.Emin, ctx.Emax, ctx.clamp, r, fl))
    except Exception as e:
        out.append("%d %s %s %s -> EXC %s" % (i, op, a, b, type(e).__name__))
for i in range(300):
    ctx = Context(prec=random.choice([1, 3, 6, 10, 28]), rounding=random.choice(modes), traps=[])
    a = rnd()
    op = random.choice(["sqrt", "exp", "ln", "log10", "logb", "normalize", "plus", "minus", "to_int", "next_plus", "next_minus", "abs", "round"])
    try:
        if op == "sqrt": r = ctx.sqrt(a)
        elif op == "exp":
            if abs(a) > 30: a = a / 1000
            r = ctx.exp(a)
        elif op == "ln": r = ctx.ln(a)
        elif op == "log10": r = ctx.log10(a)
        elif op == "logb": r = ctx.logb(a)
        elif op == "normalize": r = ctx.normalize(a)
        elif op == "plus": r = ctx.plus(a)
        elif op == "minus": r = ctx.minus(a)
        elif op == "to_int": r = ctx.to_integral_value(a)
        elif op == "next_plus": r = ctx.next_plus(a)
        elif op == "next_minus": r = ctx.next_minus(a)
        elif op == "abs": r = ctx.abs(a)
        else: r = ctx.create_decimal(a)
        out.append("u%d %s %s -> %s" % (i, op, a, r))
    except Exception as e:
        out.append("u%d %s %s -> EXC %s" % (i, op, a, type(e).__name__))
print("\n".join(out))

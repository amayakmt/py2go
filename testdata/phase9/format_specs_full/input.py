n = 255
print(f"{n:x} {n:X} {n:o} {n:b} {n:#x} {n:#o} {n:#b} {n:08b} {n:#010x}")
print(f"{n:,} {1234567:,} {1234567:_} {n:_x} {-n:x} {-n:#x}")
print(f"{n:e} {n:E} {n:g} {n:%} {n:.1%} {n:10.3e}|")
f = 1234.5678
print(f"{f:f} {f:F} {f:.2f} {f:e} {f:.3e} {f:E} {f:g} {f:G} {f:.3g} {f:%} {f:.1%}")
print(f"{f:10.2f}| {f:<10.2f}| {f:^10.2f}| {f:>+10.2f}| {f:010.2f}| {-f:010.2f}")
print(f"{f:,.2f} {f:_.1f} {1e10:g} {1e-5:g} {0.0001:g} {100000:g} {1000000:g} {1e16:g} {0.1:g}")
print(f"{12345678.9:,.1f} {0.5:.0%} {2.5:.0f} {3.5:.0f} {-0.0:.1f} {float('inf'):f} {float('nan'):.2f}")
s = "text"
print(f"{s:>8}|{s:<8}|{s:^8}|{s:*^9}|{s:.2}|{s:8.3}|")
print(f"{True:d} {True} {False:>6} {None!s:>6} {s!r:>8}")
def show(v, spec):
    return format(v, spec)
for v, spec in ((42, "05d"), (42, "x"), (3.14159, ".2f"), ("a", "^5"), (255, "#b"), (1e6, ",.0f"), (7, "+")):
    print(repr(spec), "->", show(v, spec))
try:
    format("a", "d")
except ValueError as e:
    print("ValueError", e)
try:
    format(1.5, "x")
except ValueError as e:
    print("ValueError", e)

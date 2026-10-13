vals = [0.0, 1.0, -1.5, 3.14159265358979, 1234.5678, 1e-7, 123456789.0, 2.5, 0.125, 1e100, -0.0, 0.000123456]
for v in vals:
    print("%f|%.2f|%10.3f|%-10.1f|%e|%.3E|%g|%G|%.3g|%+.1f|% f" % (v, v, v, v, v, v, v, v, v, v, v))
    print(f"{v:f} {v:.0f} {v:.10f} {v:e} {v:.0e} {v:g} {v:.2g} {v:10.2f} {v:<12.3e}| {v:^12.2f} {v:+.2f} {v:012.3f} {v:,.2f} {v:_.1f} {v:%} {v:.1%} {v:#.0f} {v:#g}")
print(format(1e16, "g"), format(1e-5, "g"), format(123456.0, "g"), format(1234567.0, "g"), format(0.0001, "g"), format(2.5, ".0f"), format(3.5, ".0f"), format(-2.5, ".0f"), format(0.5, ".0f"))
print(format(float("inf"), "f"), format(float("nan"), ".2f"), format(float("-inf"), "10.1f"), format(float("inf"), "010"), "%f" % float("nan"), "%5.1f|" % float("inf"))
print(f"{1e300:f}"[:20], f"{12345678.9:,}", f"{1234.5:,.1f}", f"{0.5:=+8.2f}", f"{-0.5:=+8.2f}", f"{3:.2f}", f"{3:e}", f"{True:.1f}")

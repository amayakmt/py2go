for n in range(-30, 31, 5):
    print(n, round(n, -1), round(n, -2), round(n / 10), round(n / 10, 0), round(n / 100, 1), round(n / 1000, 2), round(n * 10 ** 25 + 5, -26) if False else 0)
vals = [0.5, 1.5, 2.5, 3.5, 4.5, -0.5, -1.5, -2.5, 0.05, 0.15, 0.25, 0.35, 0.45, 2.675, 1.005, 1.115, 1.125, 1.135, 10.075, 8.325, 5.055, 1e-10, 123456.789, 0.1 + 0.2]
for v in vals:
    print(v, [round(v, d) for d in range(-2, 5)], round(v), int(round(v)), type(round(v)).__name__, type(round(v, 1)).__name__)
big = [2 ** 70 + k for k in (0, 5, 50, 500, -5, -50)]
print([round(b, -1) for b in big], [round(b, -2) for b in big], [round(b, -20) for b in big], [round(b, 5) for b in big], [round(float(b), -3) for b in big])
for f in (lambda: round(1.5, 1.5), lambda: round("1"), lambda: round(None), lambda: round(1, None), lambda: round(1.5, None), lambda: round(float("nan"), 2), lambda: round(float("inf"), 2), lambda: round(2.5, 400), lambda: round(2.5, -400), lambda: round(10 ** 30, -400)):
    try:
        print(f())
    except (TypeError, ValueError, OverflowError) as e:
        print(type(e).__name__, e)

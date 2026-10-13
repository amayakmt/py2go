def series_e(n):
    t, s = 1.0, 1.0
    for k in range(1, n):
        t /= k
        s += t
    return s

def simpson(f, a, b, n=1000):
    h = (b - a) / n
    s = f(a) + f(b)
    for i in range(1, n):
        s += f(a + i * h) * (4 if i % 2 else 2)
    return s * h / 3

def newton(f, df, x, iters=40):
    for _ in range(iters):
        x -= f(x) / df(x)
    return x

def leibniz(n):
    return 4 * sum((-1) ** k / (2 * k + 1) for k in range(n))

print(series_e(20), series_e(5), simpson(lambda x: x * x, 0, 3), simpson(lambda x: x ** 3 - 2 * x, -1, 2), newton(lambda x: x * x - 2, lambda x: 2 * x, 1.0), newton(lambda x: x ** 3 - 27, lambda x: 3 * x * x, 5.0))
print(leibniz(1000), leibniz(10), sum(1 / k ** 2 for k in range(1, 1001)), sum(1 / k for k in range(1, 101)), 1 / 3 + 1 / 3 + 1 / 3, sum([0.1] * 100), sum([1e16, 1.0, -1e16]))
xs = [i / 10 for i in range(-20, 21, 5)]
print(xs, [x * x for x in xs], [round(x / 3, 4) for x in xs], [x // 0.3 for x in xs], [x % 0.7 for x in xs], [abs(x) ** 0.5 for x in xs], max(xs), min(xs), sorted(xs, key=abs))
print(0.1 * 3, 0.1 + 0.7, 1.1 * 1.1, 4.35 * 100, 1e308 * 10, -1e308 * 10, 1e-320 / 1e10, 5e-324 / 2, 2.0 ** 1023 * 2, (2.0 ** 1023) * 1.9999999999999998)

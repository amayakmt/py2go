for base, exp in ((2, 0), (2, 1), (2, 10), (2, 100), (3, 40), (-2, 63), (-2, 64), (10, 25), (7, 1), (0, 5), (1, 10 ** 6), (-1, 10 ** 6), (-1, 10 ** 6 + 1)):
    print(base, exp, base ** exp)
for base, exp in ((2, -1), (2, -10), (10, -3), (-2, -3), (3, -2)):
    print(base, exp, base ** exp)
print(pow(2, 10, 1000), pow(2, 1000, 10 ** 9 + 7), pow(3, 10 ** 18, 10 ** 9 + 7), pow(-2, 3, 5), pow(2, 3, -5), pow(0, 0, 5), pow(7, 0, 1), pow(5, 3, 1))
print(pow(3, -1, 11), pow(2, -3, 7), pow(10, -1, 17), pow(123456789, -1, 10 ** 9 + 7), pow(7, -2, 13))
for args in ((2, -1, 4), (0, -1, 5), (4, -1, 6), (2, 3, 0), (2.5, 2, 3), (2, 3.0, 3)):
    try:
        print(pow(*args))
    except (ValueError, TypeError, ZeroDivisionError) as e:
        print(type(e).__name__, e)
p = 2 ** 127 - 1
print(pow(3, p - 1, p), pow(2, p - 1, p) == 1, pow(5, p, p), (p * p) % 1000003, p ** 2 - (p - 1) * (p + 1))
x = 2
for _ in range(10):
    x = x * x % (10 ** 18 + 9)
print(x)

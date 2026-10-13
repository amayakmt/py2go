for n in (0, 1, 31, 32, 62, 63, 64, 65, 100, 1000):
    print(n, 1 << n, -1 << n, (1 << n) >> n, (2 ** 70) >> n, -(2 ** 70) >> n, 5 >> n, -5 >> n)
for bad in (-1, -100):
    for f in (lambda: 1 << bad, lambda: 1 >> bad):
        try:
            print(f())
        except ValueError as e:
            print("ValueError", e)
print(0 << 10 ** 6, 0 >> 10 ** 6, -1 >> 10 ** 6, 1 >> 10 ** 6)
x = 1
for i in range(10):
    x = (x << 7) ^ (x >> 3) ^ (i * 2654435761)
    x &= (1 << 96) - 1
    print(x)

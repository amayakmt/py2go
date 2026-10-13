def sieve(n):
    flags = [True] * (n + 1)
    flags[0:2] = [False, False]
    for i in range(2, int(n ** 0.5) + 1):
        if flags[i]:
            flags[i * i::i] = [False] * len(flags[i * i::i])
    return [i for i, f in enumerate(flags) if f]

def factor(n):
    out = {}
    d = 2
    while d * d <= n:
        while n % d == 0:
            out[d] = out.get(d, 0) + 1
            n //= d
        d += 1 if d == 2 else 2
    if n > 1:
        out[n] = out.get(n, 0) + 1
    return out

ps = sieve(100)
print(ps, len(sieve(10000)), sieve(10000)[-3:])
for n in (360, 2 ** 20, 600851475143, 10 ** 12 + 39, 999999000001, 2 ** 31 - 1, 1234567890123):
    f = factor(n)
    prod = 1
    for p, k in f.items():
        prod *= p ** k
    print(n, f, prod == n)
def pollard(n):
    if n % 2 == 0:
        return 2
    x = y = 2
    c = 1
    d = 1
    while d == 1:
        x = (x * x + c) % n
        y = (y * y + c) % n
        y = (y * y + c) % n
        a = abs(x - y)
        g = n
        b = a
        while b:
            g, b = b, g % b
        d = g
    return d if d != n else None
for n in (8051, 10403, 1000000016000000063, (2 ** 31 - 1) * (2 ** 61 - 1) if False else 91):
    print(n, pollard(n))
print(sum(ps), sum(p * p for p in ps) % 1000, max(ps) ** 3, ps[10:15])

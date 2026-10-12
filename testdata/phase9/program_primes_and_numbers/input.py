def sieve(n):
    flags = [True] * (n + 1)
    flags[0:2] = [False, False]
    for i in range(2, int(n ** 0.5) + 1):
        if flags[i]:
            for j in range(i * i, n + 1, i):
                flags[j] = False
    return [i for i, f in enumerate(flags) if f]

def factorize(n):
    factors = {}
    d = 2
    while d * d <= n:
        while n % d == 0:
            factors[d] = factors.get(d, 0) + 1
            n //= d
        d += 1
    if n > 1:
        factors[n] = factors.get(n, 0) + 1
    return factors

def gcd(a, b):
    while b:
        a, b = b, a % b
    return a

def collatz(n):
    steps = 0
    while n != 1:
        n = n // 2 if n % 2 == 0 else 3 * n + 1
        steps += 1
    return steps

print(sieve(50), len(sieve(1000)))
print(factorize(360), factorize(97), factorize(1001), factorize(2 ** 10))
print(gcd(48, 18), gcd(17, 5), gcd(0, 9), [collatz(n) for n in (1, 7, 27)])
perfect = [n for n in range(2, 500) if sum(d for d in range(1, n) if n % d == 0) == n]
print(perfect, sum(range(101)), 2 ** 62, max(range(10), key=lambda n: (n % 4, n)))

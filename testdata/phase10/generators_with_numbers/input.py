def fibs():
    a, b = 0, 1
    while True:
        yield a
        a, b = b, a + b

def take(it, n):
    out = []
    for x in it:
        if len(out) == n:
            break
        out.append(x)
    return out

def primes():
    found = []
    n = 2
    while True:
        if all(n % p for p in found if p * p <= n):
            found.append(n)
            yield n
        n += 1

def collatz(n):
    while n != 1:
        yield n
        n = n // 2 if n % 2 == 0 else 3 * n + 1
    yield 1

def powers(base):
    x = 1
    while True:
        yield x
        x *= base

print(take(fibs(), 12), take(fibs(), 100)[-1], take(primes(), 15), list(collatz(27))[:12], len(list(collatz(2 ** 70 + 1))), take(powers(2), 70)[-1], take(powers(10), 25)[-1], take(powers(-3), 5))
g = fibs()
for _ in range(90):
    next(g)
print(next(g), next(g), sum(take(fibs(), 100)), max(take(fibs(), 100)) == take(fibs(), 100)[-1])
import_zip = list(zip(powers(2), "abcdef", fibs()))
print(import_zip, dict(zip("abc", powers(7))), sum(x for x, _ in zip(powers(3), range(10))), [x for x in take(powers(2), 10) if x % 3 == 1], list(map(lambda v: v // 1000, take(powers(10), 8))), any(x > 10 ** 20 for x in powers(10)), next(x for x in fibs() if x > 10 ** 30))

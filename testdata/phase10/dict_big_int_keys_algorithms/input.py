memo = {}
def fib(n):
    if n < 2:
        return n
    if n in memo:
        return memo[n]
    memo[n] = fib(n - 1) + fib(n - 2)
    return memo[n]
print(fib(90), len(memo), fib(300) % 10 ** 9)
counts = {}
for i in range(1, 60):
    key = 2 ** i
    counts[key] = counts.get(key, 0) + i
print(len(counts), counts[2 ** 59], sorted(counts)[-2:], 2 ** 62 in counts, 2 ** 64 in counts)
rev = {v: k for k, v in counts.items()}
print(len(rev), rev[1], rev[59] == 2 ** 59, sorted(rev)[:5])
groups = {}
for n in range(2 ** 60 - 5, 2 ** 60 + 5):
    groups.setdefault(n % 3, []).append(n)
print({k: len(v) for k, v in sorted(groups.items())}, groups[0][:2], groups[1][-1])
d = dict.fromkeys([2 ** 70, 2 ** 71, 3], 0)
d[2 ** 70] += 5
d[2 ** 71] -= 2 ** 80
print(d, d.pop(3), d.get(2 ** 72, "none"), list(d), list(d.values()), max(d, key=d.get), sum(d.values()))
cache = {}
def collatz(n):
    if n == 1:
        return 0
    if n in cache:
        return cache[n]
    r = 1 + collatz(n // 2 if n % 2 == 0 else 3 * n + 1)
    cache[n] = r
    return r
print(collatz(2 ** 70 + 1), collatz(837799), len(cache) > 100, max(cache.values()), min(cache, key=cache.get) if False else 0)

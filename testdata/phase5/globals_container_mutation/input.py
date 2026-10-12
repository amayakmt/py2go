cache: dict[int, int] = {}
seen: set[int] = set()
log: list[str] = []


def fib(n: int) -> int:
    if n < 2:
        return n
    if n in cache:
        return cache[n]
    seen.add(n)
    log.append("fib" + str(n))
    cache[n] = fib(n - 1) + fib(n - 2)
    return cache[n]


print(fib(20), len(cache), len(seen), len(log))
print(fib(20), len(log))

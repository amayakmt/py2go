calls: list[int] = [0]


def check(n: int, calls: list[int]) -> bool:
    calls[0] += 1
    print("check", n)
    return n > 2


print(any(check(n, calls) for n in range(10)))
print(calls)
print(all(check(n, calls) for n in range(10)))
print(calls)
print(any(x > 100 for x in range(5)), all(x < 100 for x in range(5)))
print(any(x for x in range(0)), all(x for x in range(0)))

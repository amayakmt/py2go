base: int = 10
count: list[int] = [0]


def next_id() -> int:
    count[0] += 1
    return count[0]


def f(a: int = base * 2, b: int = next_id(), c: list[int] = [1, 2]) -> int:
    c.append(a)
    return a + b + len(c)


base = 99
print(f(), f(), f(1))
print(count[0])
print(f(c=[7]))

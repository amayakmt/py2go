def total(xs: tuple[int, ...]) -> int:
    t: int = 0
    for x in xs:
        t += x
    return t


def make(n: int) -> tuple[int, ...]:
    return tuple(range(n))


print(total((1, 2, 3)), total(()), total(make(5)))
m: tuple[int, ...] = make(3) + (10, 20)
print(m, len(m))
nested: list[tuple[int, ...]] = [(1,), (1, 2), ()]
print(nested, sorted(nested))
try:
    print(m[10])
except IndexError as e:
    print(e)

def total(*xs: int) -> int:
    return sum(xs)


def show(*xs: str) -> str:
    return "|".join(xs)


def count(*xs: int) -> int:
    return len(xs)


print(total(), total(1), total(1, 2, 3), show("a", "b"), count(), count(1, 2))
print(total(*[1, 2], *(3, 4)), total(*range(5)), total(*{1, 2}))


def first_and_rest(first: int, *rest: int) -> tuple[int, int]:
    return first, len(rest)


print(first_and_rest(1), first_and_rest(1, 2, 3))


def inner_use(*xs: int) -> list[int]:
    out: list[int] = []
    for i, x in enumerate(xs):
        out.append(i * x)
    return out


print(inner_use(5, 6, 7), inner_use())

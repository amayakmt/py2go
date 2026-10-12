a: list[int] = [1, 2]
b: list[int] = [3, 4]
print([*a, *b], [0, *a, 9, *b], [*a, *a])
print((*a, 5), (*a, *b))
s: set[int] = {*a, *b, 1}
print(sorted(s))
d1: dict[str, int] = {"x": 1}
d2: dict[str, int] = {"y": 2, "x": 3}
print({**d1, **d2}, {**d1, "z": 9})
print([*range(3)], [*"ab"], [*d2], sorted({*d1, *d2}))
print(*a, *b, sep="-")


def total(*nums: int) -> int:
    return sum(nums)


print(total(*a, *b), total(*a, 10, *b))

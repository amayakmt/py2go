def common(a: list[int], b: list[int]) -> list[int]:
    return sorted(set(a) & set(b))


def has_duplicates(xs: list[int]) -> bool:
    return len(set(xs)) != len(xs)


def collect(n: int, out: set[int]) -> None:
    for i in range(n):
        out.add(i % 4)


print(common([1, 2, 3, 4], [3, 4, 5]), has_duplicates([1, 2, 2]), has_duplicates([1, 2, 3]))
acc: set[int] = set()
collect(10, acc)
print(acc)

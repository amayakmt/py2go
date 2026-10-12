def min_max(xs: list[int]) -> tuple[int, int]:
    return min(xs), max(xs)


def stats(xs: list[float]) -> tuple[float, float, int]:
    total: float = sum(xs)
    return total, total / len(xs), len(xs)


def swap(a: int, b: int) -> tuple[int, int]:
    return b, a


lo, hi = min_max([3, 9, 1])
print(lo, hi)
s, mean, n = stats([1.0, 2.0, 4.5])
print(s, mean, n)
print(swap(1, 2), swap(3, 4))
r: tuple[int, int] = min_max([5])
print(r, r[0] + r[1])

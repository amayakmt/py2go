def mean(xs: list[float]) -> float:
    return sum(xs) / len(xs)


def variance(xs: list[float]) -> float:
    m: float = mean(xs)
    total: float = 0.0
    for x in xs:
        total += (x - m) ** 2
    return total / len(xs)


data: list[float] = [2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0]
print(f"mean={mean(data):.3f} var={variance(data):.3f} sd={variance(data) ** 0.5:.3f}")
print(sorted(data, reverse=True)[:3])
print(min(data), max(data), max(data) - min(data))

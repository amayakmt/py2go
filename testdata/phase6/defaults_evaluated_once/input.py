def append_to(x: int, items: list[int] = []) -> list[int]:
    items.append(x)
    return items


print(append_to(1))
print(append_to(2))
print(append_to(3, [9]))
print(append_to(4))


def tally(key: str, counts: dict[str, int] = {}) -> int:
    counts[key] = counts.get(key, 0) + 1
    return counts[key]


print(tally("a"), tally("a"), tally("b"), tally("a"))

offset: int = 10


def shifted(x: int, by: int = offset) -> int:
    return x + by


offset = 1000
print(shifted(1), shifted(1, 5))

def scale(xs: list[int], k: int) -> list[int]:
    return [x * k for x in xs]


def squares_below(n: int) -> dict[int, int]:
    return {i: i * i for i in range(n)}


def evens(xs: list[int]) -> set[int]:
    return {x for x in xs if x % 2 == 0}


print(scale([1, 2, 3], 10), squares_below(4), evens([1, 2, 4, 6, 4]))
names: list[str] = ["ann", "bob"]
greetings: list[str] = [f"hi {n}" for n in names]
print(greetings)
lookup: dict[str, int] = {"a": 1, "b": 2}
print([lookup[k] * 2 for k in lookup], [k for k, v in lookup.items() if v > 1])

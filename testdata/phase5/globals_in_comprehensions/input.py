FACTOR: int = 3
BASE: list[int] = [1, 2, 3]
LOOKUP: dict[str, int] = {"a": 10}


def scaled() -> list[int]:
    return [x * FACTOR for x in BASE]


def keyed() -> dict[str, int]:
    return {k: v + FACTOR for k, v in LOOKUP.items()}


print(scaled(), keyed(), [x + FACTOR for x in BASE], sum(x for x in BASE if x > FACTOR - 2))

LIMIT: int = 10
NAMES: list[str] = ["a", "b"]
SCALE: float = 1.5
GREETING: str = "hi"
TABLE: dict[str, int] = {"x": 1}


def over(n: int) -> bool:
    return n > LIMIT


def describe() -> str:
    return GREETING + " " + ",".join(NAMES) + " " + str(SCALE * 2) + " " + str(TABLE["x"])


print(over(5), over(50), describe())

MAX: int = 10
RATE: float = 0.5
NAME: str = "limits"
TAGS: list[str] = ["a", "b"]
LOOKUP: dict[str, int] = {"x": 1}

def clamp(v: int) -> int:
    return min(v, MAX)

def scaled(v: float) -> float:
    return v * RATE

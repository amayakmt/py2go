class Color:
    RED: int = 1
    GREEN: int = 2
    BLUE: int = 3


LIMIT: int = 10


def name(c: int) -> str:
    match c:
        case Color.RED:
            return "red"
        case Color.GREEN | Color.BLUE:
            return "cool"
        case _:
            return "unknown"


def size(n: int) -> str:
    match n:
        case 0:
            return "none"
        case n if n < LIMIT:
            return "small"
        case _:
            return "large"


print([name(c) for c in range(5)], [size(n) for n in [0, 3, 10, 99]])

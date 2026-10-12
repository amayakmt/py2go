def add(d: dict[str, int], k: str, v: int) -> None:
    d[k] = v


def total(d: dict[str, int]) -> int:
    t: int = 0
    for v in d.values():
        t += v
    return t


def invert(d: dict[str, int]) -> dict[int, str]:
    out: dict[int, str] = {}
    for k, v in d.items():
        out[v] = k
    return out


data: dict[str, int] = {}
add(data, "a", 1)
add(data, "b", 2)
print(data, total(data), invert(data))
print(total({}), invert({}))

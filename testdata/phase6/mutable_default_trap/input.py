def add(x: int, acc: list[int] = []) -> list[int]:
    acc.append(x)
    return acc


print(add(1))
print(add(2))
print(add(3, []))
print(add(4))


def put(k: str, v: int, d: dict[str, int] = {}) -> dict[str, int]:
    d[k] = v
    return d


print(put("a", 1), put("b", 2), put("c", 3, {}))

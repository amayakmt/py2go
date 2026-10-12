d: dict[str, int] = {"a": 1}
d["a"] += 10
d["a"] *= 2
d["a"] -= 1
d["a"] //= 3
print(d)
f: dict[int, float] = {1: 1.5}
f[1] += 1
f[1] /= 2
print(f)
s: dict[str, str] = {"k": "x"}
s["k"] += "y"
s["k"] *= 2
print(s)
calls: list[int] = [0]


def key(counter: list[int]) -> str:
    counter[0] += 1
    return "a"


d[key(calls)] += 1
print(d, calls)

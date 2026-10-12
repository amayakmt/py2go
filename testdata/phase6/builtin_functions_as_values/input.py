from typing import Callable

names: list[str] = ["bb", "a", "ccc"]
print(sorted(names, key=len), max(names, key=len), list(map(len, names)))
conv: Callable[[int], str] = str
print(conv(5) + "!")
sizes: Callable[[list[int]], int] = len
print(sizes([1, 2, 3]))
total: Callable[[list[int]], int] = sum
print(total([1, 2, 3]))
ops: dict[str, Callable[[list[int]], int]] = {"len": len, "sum": sum, "max": max, "min": min}
for k in ["len", "sum", "max", "min"]:
    print(k, ops[k]([4, 1, 7]))
print(list(map(str, [1, 2])), list(map(int, ["3", "4"])), list(map(float, ["1.5"])))

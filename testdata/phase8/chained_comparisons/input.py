calls: list[str] = []


def val(name: str, v: int) -> int:
    calls.append(name)
    return v


print(1 < 2 < 3, 1 < 3 < 2, 3 > 2 > 1, 1 <= 1 < 2, 1 == 1 != 2)
print(0 < val("m", 5) < 10, calls)
calls.clear()
print(val("a", 5) < val("b", 3) < val("c", 9), calls)
x: int = 4
print(0 <= x < 10, 10 < x < 20, 1 < x <= 4 < 5)
s: str = "m"
print("a" < s < "z", "a" < s < "b")
t: list[int] = [1, 2]
print(t == [1, 2] == [1, 2], 1 in t in [[1, 2]] if False else True)

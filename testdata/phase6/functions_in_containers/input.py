from typing import Callable


def add(a: int, b: int) -> int:
    return a + b


def mul(a: int, b: int) -> int:
    return a * b


ops: dict[str, Callable[[int, int], int]] = {"add": add, "mul": mul}
print(ops["add"](3, 4), ops["mul"](3, 4))
for name in ["add", "mul"]:
    print(name, ops[name](5, 6))
table: list[Callable[[int, int], int]] = [add, mul, lambda a, b: a - b]
print([f(10, 3) for f in table])
pair: tuple[Callable[[int, int], int], str] = (mul, "times")
print(pair[0](6, 7), pair[1])
ops["sub"] = lambda a, b: a - b
print(sorted(ops), ops["sub"](9, 4))

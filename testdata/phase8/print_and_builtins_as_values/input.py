from typing import Callable

names: list[str] = ["a", "b"]
for name_ in names:
    print(name_)
emit: Callable[[str], None] = print
emit("through a variable")
show: Callable[[int, int], None] = lambda a, b: print(a + b)
show(1, 2)
handlers: dict[str, Callable[[str], None]] = {"p": print}
handlers["p"]("from dict")

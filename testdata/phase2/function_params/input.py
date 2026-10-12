def add(a: int, b: int) -> int:
    return a + b


def scale(x: float, k: float) -> float:
    return x * k


def greet(name: str, excited: bool) -> str:
    if excited:
        return "Hello, " + name + "!"
    return "Hello, " + name


print(add(2, 3))
print(add(add(1, 2), add(3, 4)))
print(scale(1.5, 2.0))
print(greet("Ada", True))
print(greet("Bob", False))

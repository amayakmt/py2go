def show(label: str, v: int) -> int:
    print("eval", label)
    return v


def f(a: int, b: int, c: int) -> int:
    return a * 100 + b * 10 + c


print(f(show("a", 1), show("b", 2), show("c", 3)))
print(f(c=show("c", 3), a=show("a", 1), b=show("b", 2)))
print(f(show("a", 1), c=show("c", 3), b=show("b", 2)))

trace: list[str] = []


def t(name: str, v: int) -> int:
    trace.append(name)
    return v


def f(a: int, b: int, c: int = 0, *, d: int = 0) -> int:
    return a + b + c + d


print(f(t("a", 1), t("b", 2), d=t("d", 4)))
print(f(b=t("b", 2), a=t("a", 1), c=t("c", 3)))
print(trace)
trace.clear()
print(f(*[t("x", 1), t("y", 2)]))
print(trace)

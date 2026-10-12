def f(a: int, b: int = 0) -> None:
    pass


d: dict[str, int] = {"b": 1}
f(a=1, **d)

def f() -> None:
    try:
        raise ValueError("a")
    finally:
        raise KeyError("b")


f()

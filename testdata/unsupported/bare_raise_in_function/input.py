def f() -> None:
    raise


try:
    pass
except ValueError:
    f()

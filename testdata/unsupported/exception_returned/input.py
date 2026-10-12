def f() -> int:
    try:
        return 1
    except ValueError as e:
        return e

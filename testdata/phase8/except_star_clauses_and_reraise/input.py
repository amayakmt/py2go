def f(kind: str) -> None:
    try:
        if kind == "mixed":
            raise ExceptionGroup("mix", [ValueError("v"), KeyError("k"), TypeError("t")])
        if kind == "single":
            raise KeyError("only")
        if kind == "nested":
            raise ExceptionGroup("outer", [ExceptionGroup("inner", [ValueError("a"), OSError("o")]), ValueError("b")])
    except* ValueError as eg:
        print("V", [str(e) for e in eg.exceptions])
    except* (KeyError, TypeError) as eg:
        print("K/T", sorted(type(e).__name__ for e in eg.exceptions))


for k in ["mixed", "single", "nested", "none"]:
    try:
        f(k)
        print(k, "handled")
    except BaseException as e:
        print(k, "escaped", type(e).__name__, str(e))

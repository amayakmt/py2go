def run(kind: int) -> None:
    try:
        if kind == 0:
            raise ExceptionGroup("g", [ValueError("v1"), TypeError("t1"), ValueError("v2")])
        if kind == 1:
            raise ValueError("lonely")
        if kind == 2:
            raise ExceptionGroup("g2", [KeyError("k")])
        print("no error")
    except* ValueError as e:
        print("values:", e, [str(x) for x in e.exceptions])
    except* TypeError as e:
        print("types:", e)
    else:
        print("else ran")
    finally:
        print("finally", kind)


for k in range(4):
    try:
        run(k)
    except ExceptionGroup as rest:
        print("leftover:", rest, [type(x).__name__ for x in rest.exceptions])
    except Exception as other:
        print("plain:", other)

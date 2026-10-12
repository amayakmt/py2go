def c() -> None:
    try:
        print("c start")
        raise ValueError("deep")
    finally:
        print("c finally")


def b() -> None:
    try:
        c()
    finally:
        print("b finally")


def a() -> None:
    try:
        b()
    except ValueError as e:
        print("a caught", e)
    finally:
        print("a finally")


a()

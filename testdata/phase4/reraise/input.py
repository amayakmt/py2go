def log_and_raise(n: int) -> None:
    try:
        if n > 1:
            raise ValueError("too big")
        print("ok", n)
    except ValueError as e:
        print("logging", e)
        raise e


for k in range(3):
    try:
        log_and_raise(k)
    except ValueError as e:
        print("caller saw", repr(e))

try:
    raise ValueError("bad value", 3)
except ValueError as e:
    print(e)
    print(str(e), repr(e))
    print(f"{e}|{e!r}|{e!s}")
    print("msg: " + str(e), len(str(e)), "bad" in str(e))
    print(str(e).upper())
    if e:
        print("always truthy")
    print(not e)
try:
    raise RuntimeError
except RuntimeError as e:
    print(repr(str(e)), len(str(e)))

def convert(s: str) -> int:
    try:
        return int(s)
    except ValueError as e:
        raise RuntimeError("cannot convert " + s) from e


def quiet(s: str) -> int:
    try:
        return int(s)
    except ValueError:
        raise RuntimeError("quiet") from None


try:
    print(convert("5"))
    print(convert("five"))
except RuntimeError as e:
    print(e)
try:
    quiet("x")
except RuntimeError as e:
    print(e)

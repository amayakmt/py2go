class Low(Exception):
    pass


class High(Exception):
    pass


def low() -> None:
    raise Low("disk")


def high() -> None:
    try:
        low()
    except Low as e:
        raise High("service") from e


try:
    high()
except High as e:
    print(e, type(e.__cause__).__name__, e.__cause__)
    print(e.__context__ is e.__cause__, e.__suppress_context__)

try:
    try:
        raise ValueError("inner")
    except ValueError:
        raise KeyError("outer")
except KeyError as e:
    print(type(e.__context__).__name__, e.__cause__ is None, e.__suppress_context__)

try:
    try:
        raise ValueError("inner")
    except ValueError:
        raise RuntimeError("quiet") from None
except RuntimeError as e:
    print(e.__cause__ is None, e.__context__ is None, e.__suppress_context__)

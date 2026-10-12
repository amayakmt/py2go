def inner() -> None:
    raise TypeError("from inner")


def middle() -> None:
    try:
        inner()
    except ValueError:
        print("middle: wrong class")


def outer() -> None:
    try:
        middle()
    except TypeError as e:
        print("outer caught", e)


outer()
try:
    middle()
except Exception as e:
    print("top caught", repr(e))

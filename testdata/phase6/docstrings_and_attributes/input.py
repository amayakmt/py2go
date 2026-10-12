def documented(x: int) -> int:
    """Return x plus one.

    Longer text here.
    """
    return x + 1


def plain(x: int) -> int:
    return x


def outer() -> str:
    def inner() -> str:
        "inner doc"
        return "x"
    print(inner.__qualname__, inner.__name__, inner.__doc__)
    return inner()


print(documented.__name__, documented.__qualname__)
print(repr(documented.__doc__))
print(plain.__name__)
print(outer())
print(__name__)
"a string statement on its own is not a docstring and does nothing"
42
...

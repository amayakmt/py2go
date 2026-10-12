def show(label: str) -> None:
    print(label)


try:
    raise ValueError
except ValueError as e:
    print(repr(e), repr(str(e)))
try:
    raise ValueError()
except ValueError as e:
    print(repr(e), repr(str(e)))
try:
    raise ValueError("one")
except ValueError as e:
    print(repr(e), repr(str(e)))
try:
    raise ValueError("two", 2)
except ValueError as e:
    print(repr(e), repr(str(e)))
try:
    raise ValueError(42)
except ValueError as e:
    print(repr(e), repr(str(e)))
try:
    raise TypeError(1.5, True, "x")
except TypeError as e:
    print(repr(e), repr(str(e)))
try:
    raise ValueError([1, 2])
except ValueError as e:
    print(repr(e), repr(str(e)))
try:
    raise ValueError("it's")
except ValueError as e:
    print(repr(e), repr(str(e)))
try:
    raise KeyError("k")
except KeyError as e:
    print(repr(e), repr(str(e)), e)

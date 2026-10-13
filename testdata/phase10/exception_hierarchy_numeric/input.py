for exc in (ZeroDivisionError, OverflowError, FloatingPointError, ArithmeticError, ValueError, UnicodeDecodeError, UnicodeEncodeError, UnicodeError, LookupError, IndexError, KeyError, TypeError, AttributeError, OSError, RecursionError, MemoryError, StopIteration, AssertionError):
    print(exc.__name__, issubclass(exc, Exception), issubclass(exc, ArithmeticError), issubclass(exc, ValueError), issubclass(exc, LookupError))
def classify(f):
    try:
        f()
    except ArithmeticError as e:
        return "arith:" + type(e).__name__
    except UnicodeError as e:
        return "unicode:" + type(e).__name__
    except ValueError as e:
        return "value"
    except LookupError as e:
        return "lookup:" + type(e).__name__
    except TypeError:
        return "type"
    except Exception as e:
        return "other:" + type(e).__name__
    return "none"
cases = [lambda: 1 / 0, lambda: 10 ** 5000 * 1.0, lambda: int("z"), lambda: b"\xff".decode(), lambda: "é".encode("ascii"), lambda: [][0], lambda: {}[1], lambda: 1 + "a", lambda: float("1e999") * 0 + int(float("inf")), lambda: chr(-1), lambda: None.x, lambda: 5, lambda: (1).bit_length(2), lambda: bytes(-1), lambda: "a".encode("bogus")]
print([classify(c) for c in cases])
try:
    try:
        1 / 0
    except ZeroDivisionError as e:
        raise ValueError("wrapped") from e
except ValueError as e:
    print(e, type(e.__cause__).__name__, e.__cause__)
class NumErr(ArithmeticError):
    pass
try:
    raise NumErr("custom", 42)
except ArithmeticError as e:
    print(type(e).__name__, e.args, str(e), repr(e))

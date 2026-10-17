class AppError(Exception):
    pass


def attempt(fn, exc, default):
    try:
        return fn()
    except exc as e:
        print("caught", type(e).__name__, e)
        return default


print(attempt(lambda: int("x"), ValueError, -1))
print(attempt(lambda: {}["k"], (KeyError, IndexError), -2))
print(attempt(lambda: 1 / 0, ArithmeticError, -3))


def boom():
    raise AppError("bad")


print(attempt(boom, AppError, -4))
errors = (TypeError, AppError)
try:
    boom()
except errors:
    print("tuple variable")
try:
    attempt(lambda: [][1], KeyError, 0)
except IndexError:
    print("not caught by the wrong class")

class Wrapped(Exception):
    pass


def inner() -> None:
    raise ValueError("root cause")


try:
    inner()
except ValueError as e:
    raise Wrapped("outer") from e

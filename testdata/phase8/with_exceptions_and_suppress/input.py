class Guard:
    def __init__(self, swallow: bool) -> None:
        self.swallow = swallow

    def __enter__(self) -> int:
        return 42

    def __exit__(self, exc_type: object, exc: object, tb: object) -> bool:
        print("exit saw", exc_type, exc)
        return self.swallow


with Guard(True) as v:
    print("value", v)
    raise ValueError("swallowed")
print("continued")
try:
    with Guard(False):
        raise KeyError("kept")
except KeyError as e:
    print("caught", e)
try:
    with Guard(False) as w:
        print(w + 1)
finally:
    print("finally")

class AppError(Exception):
    pass


class NotFound(AppError):
    def __init__(self, key: str) -> None:
        super().__init__("missing " + key)
        self.key = key


class Invalid(AppError):
    def __init__(self, code: int, reason: str) -> None:
        self.code = code
        self.reason = reason

    def __str__(self) -> str:
        return f"[{self.code}] {self.reason}"


def find(d: dict[str, int], k: str) -> int:
    if k not in d:
        raise NotFound(k)
    if d[k] < 0:
        raise Invalid(d[k], "negative")
    return d[k]


data: dict[str, int] = {"a": 1, "b": -2}
for key in ["a", "b", "c"]:
    try:
        print(find(data, key))
    except NotFound as e:
        print("not found:", e, e.key, type(e).__name__)
    except Invalid as e:
        print("invalid:", e, e.code, e.reason)
try:
    raise AppError("plain", 42)
except AppError as e:
    print(e, repr(e))
try:
    find(data, "zzz")
except AppError as e:
    print("base catch:", e, isinstance(e, NotFound), isinstance(e, Exception))
print(issubclass(NotFound, AppError), issubclass(NotFound, Exception), issubclass(AppError, NotFound))

from typing import Optional


class Logger:
    def __init__(self) -> None:
        self.seen: list[str] = []

    def __enter__(self) -> "Logger":
        return self

    def __exit__(self, t: object, e: Optional[BaseException], tb: object) -> Optional[bool]:
        if e is None:
            self.seen.append("clean")
            return None
        self.seen.append("error:" + str(e))
        return True


lg: Logger = Logger()
with lg:
    pass
with lg:
    raise RuntimeError("one")
with lg:
    x: int = 1 // 0
print(lg.seen)


class Bad:
    def __enter__(self) -> None:
        pass

    def __exit__(self, t: object, e: object, tb: object) -> None:
        raise KeyError("from exit")


try:
    with Bad():
        raise ValueError("body")
except KeyError as e:
    print("exit error wins:", e, type(e.__context__).__name__)

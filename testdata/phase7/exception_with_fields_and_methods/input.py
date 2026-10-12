class HttpError(Exception):
    status: int = 500

    def __init__(self, message: str, status: int) -> None:
        super().__init__(message)
        self.message = message
        self.status = status

    def retryable(self) -> bool:
        return self.status >= 500

    def describe(self) -> str:
        return f"{self.status} {self.message}"


class Timeout(HttpError):
    def __init__(self) -> None:
        super().__init__("timeout", 504)


errors: list[HttpError] = [HttpError("bad", 400), Timeout(), HttpError("oops", 500)]
for err in errors:
    try:
        raise err
    except HttpError as e:
        print(e.describe(), e.retryable(), str(e))
print(len([e for e in errors if e.retryable()]))

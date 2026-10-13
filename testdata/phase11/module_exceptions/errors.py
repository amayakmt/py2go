class AppError(Exception):
    pass

class NotFound(AppError):
    def __init__(self, what):
        super().__init__(f"{what} not found")
        self.what = what

def find(x):
    if x < 0:
        raise NotFound(f"item {x}")
    if x == 0:
        raise AppError("zero")
    return x

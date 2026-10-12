a: int = 1
b: float = 2.0
c: str = "never read"
d: bool = True
a = 5


def f(unused: int) -> None:
    local: int = 3


f(1)
print("compiled")

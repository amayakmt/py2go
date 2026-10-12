class Money:
    def __init__(self, c: int) -> None:
        self.c = c

    def __add__(self, other: "Money") -> "Money":
        return Money(self.c + other.c)

    def __radd__(self, other: int) -> "Money":
        return Money(self.c + other)

    def __repr__(self) -> str:
        return f"Money({self.c})"


wallet: list[Money] = [Money(100), Money(250), Money(5)]
print(sum(wallet, Money(0)))
print(sum(wallet))
print(sum([m for m in wallet if m.c > 50], Money(1)))

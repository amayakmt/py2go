class Money:
    def __init__(self, cents: int) -> None:
        self.cents = cents

    def __repr__(self) -> str:
        return f"Money({self.cents})"

    def __str__(self) -> str:
        return f"${self.cents // 100}.{self.cents % 100:02d}"


class Plain:
    def __init__(self, v: int) -> None:
        self.v = v

    def __repr__(self) -> str:
        return "Plain<" + str(self.v) + ">"


m: Money = Money(1234)
print(m)
print(repr(m), str(m))
print([m, m])
print(f"{m} {m!r}")
p: Plain = Plain(1)
print(p, str(p), [p])
items: dict[str, Money] = {"a": m}
print(items)

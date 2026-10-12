class Builder:
    def __init__(self) -> None:
        self.parts: list[str] = []

    def add(self, s: str) -> "Builder":
        self.parts.append(s)
        return self

    def twice(self) -> "Builder":
        self.parts = self.parts + self.parts
        return self

    def build(self) -> str:
        return "".join(self.parts)


print(Builder().add("a").add("b").twice().add("!").build())
b: Builder = Builder()
b2: Builder = b.add("x")
print(b is b2, b.build())

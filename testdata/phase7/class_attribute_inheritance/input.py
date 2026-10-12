class Base:
    kind: str = "base"
    registry: list[str] = []

    def __init__(self, name: str) -> None:
        self.name = name
        Base.registry.append(name)

    def label(self) -> str:
        return self.kind + ":" + self.name


class Derived(Base):
    kind: str = "derived"


class Further(Derived):
    pass


items: list[Base] = [Base("a"), Derived("b"), Further("c")]
print([i.label() for i in items])
print(Base.kind, Derived.kind, Further.kind)
print(Base.registry, Derived.registry is Base.registry)
Derived.kind = "changed"
print([i.label() for i in items])

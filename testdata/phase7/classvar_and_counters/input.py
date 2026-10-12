from typing import ClassVar


class Sheep:
    total: ClassVar[int] = 0
    names: ClassVar[list[str]] = []

    def __init__(self, name: str) -> None:
        self.name = name
        Sheep.total += 1
        Sheep.names.append(name)

    @classmethod
    def flock(cls) -> str:
        return ",".join(cls.names) + ":" + str(cls.total)


Sheep("dolly")
Sheep("molly")
s: Sheep = Sheep("polly")
print(Sheep.flock(), s.flock(), Sheep.total)

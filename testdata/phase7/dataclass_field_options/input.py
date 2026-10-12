from dataclasses import dataclass, field
from typing import ClassVar


@dataclass
class Config:
    name: str
    retries: int = 3
    secret: str = field(default="hidden", repr=False)
    cache: list[int] = field(default_factory=list, compare=False)
    derived: int = field(init=False, default=0)
    count: ClassVar[int] = 0

    def __post_init__(self) -> None:
        self.derived = self.retries * 2
        Config.count += 1


a: Config = Config("a")
b: Config = Config("b", 5, "pw", [1, 2])
c: Config = Config("a")
c.cache.append(99)
print(a, b, a == c, a == b, Config.count)
print(a.derived, b.derived, b.secret, len(c.cache), len(a.cache))

from typing import Optional


class Config:
    def __init__(self, parent: Optional["Config"], key: str, value: Optional[int]) -> None:
        self.parent = parent
        self.key = key
        self.value = value

    def resolve(self) -> Optional[int]:
        if self.value is not None:
            return self.value
        if self.parent is not None:
            return self.parent.resolve()
        return None

    def depth(self) -> int:
        d: int = 0
        node: Optional[Config] = self
        while node is not None:
            d += 1
            node = node.parent
        return d


root: Config = Config(None, "root", 10)
mid: Config = Config(root, "mid", None)
leaf: Config = Config(mid, "leaf", None)
lone: Config = Config(None, "lone", None)
print(leaf.resolve(), mid.resolve(), lone.resolve())
print(leaf.depth(), root.depth())


def greet(name: Optional[str] = None, times: int = 1) -> str:
    who: str = name if name is not None else "world"
    return ("hello " + who + "! ") * times


print(greet(), greet("bob", 2), greet(times=2))

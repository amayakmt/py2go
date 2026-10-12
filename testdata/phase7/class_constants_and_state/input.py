class Config:
    MAX: int = 3
    NAMES: list[str] = ["x", "y"]
    ratio = 0.5

    def __init__(self) -> None:
        self.items: list[int] = []

    def push(self, v: int) -> bool:
        if len(self.items) >= self.MAX:
            return False
        self.items.append(v)
        return True

    def scaled(self, v: float) -> float:
        return v * Config.ratio


c: Config = Config()
print([c.push(i) for i in range(5)], c.items, Config.MAX, c.MAX)
Config.MAX = 5
print([c.push(i) for i in range(5)], c.items)
Config.NAMES.append("z")
print(Config.NAMES, c.NAMES is Config.NAMES, c.scaled(3.0))

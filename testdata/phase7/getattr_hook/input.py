class Config:
    def __init__(self) -> None:
        self.values: dict[str, int] = {"a": 1, "b": 2}
        self.direct = 10

    def __getattr__(self, name: str) -> int:
        if name in self.values:
            return self.values[name]
        raise AttributeError("no setting " + name)


c: Config = Config()
print(c.direct, c.a, c.b)
try:
    print(c.zzz)
except AttributeError as e:
    print("AttributeError:", e)

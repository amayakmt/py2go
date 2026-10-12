class Resource:
    def __init__(self, name: str) -> None:
        self.name = name
        self.open = True
        self.events: list[str] = []

    def use(self, fail: bool) -> int:
        try:
            self.events.append("use " + self.name)
            if fail:
                raise RuntimeError("boom in " + self.name)
            return 1
        finally:
            self.events.append("cleanup " + self.name)
            self.open = False


r1: Resource = Resource("a")
r2: Resource = Resource("b")
print(r1.use(False), r1.open)
try:
    r2.use(True)
except RuntimeError as e:
    print(e, r2.open)
print(r1.events, r2.events)

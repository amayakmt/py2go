class Pipeline:
    def __init__(self) -> None:
        self.steps: list[str] = []
        self.setup()
        self.result = self.compute(3)

    def setup(self) -> None:
        self.steps.append("setup")
        self.scale = 10

    def compute(self, n: int) -> int:
        self.steps.append("compute")
        return self.helper(n) + self.scale

    def helper(self, n: int) -> int:
        self.steps.append("helper")
        return n * n


p: Pipeline = Pipeline()
print(p.steps, p.result, p.scale)

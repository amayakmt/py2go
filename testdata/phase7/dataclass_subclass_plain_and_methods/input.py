from dataclasses import dataclass


@dataclass
class Shape:
    name: str

    def describe(self) -> str:
        return "shape " + self.name


class Labeled(Shape):
    def __init__(self, name: str, label: str) -> None:
        super().__init__(name)
        self.label = label

    def describe(self) -> str:
        return super().describe() + " /" + self.label


s: Shape = Labeled("sq", "big")
print(s.describe(), s)
print(Labeled("a", "x") == Labeled("a", "y"), isinstance(s, Shape))

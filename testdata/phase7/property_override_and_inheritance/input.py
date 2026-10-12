class Shape:
    @property
    def sides(self) -> int:
        return 0

    @property
    def name(self) -> str:
        return "shape" + str(self.sides)


class Triangle(Shape):
    @property
    def sides(self) -> int:
        return 3


class Square(Shape):
    @property
    def sides(self) -> int:
        return 4

    @property
    def name(self) -> str:
        return "square"


shapes: list[Shape] = [Shape(), Triangle(), Square()]
for s in shapes:
    print(s.name, s.sides)

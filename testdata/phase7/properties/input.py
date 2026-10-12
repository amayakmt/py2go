class Circle:
    def __init__(self, radius: float) -> None:
        self._radius = radius

    @property
    def radius(self) -> float:
        return self._radius

    @radius.setter
    def radius(self, value: float) -> None:
        if value < 0.0:
            raise ValueError("negative radius")
        self._radius = value

    @property
    def area(self) -> float:
        return 3.0 * self._radius * self._radius

    @property
    def diameter(self) -> float:
        return self._radius * 2.0


c: Circle = Circle(2.0)
print(c.radius, c.area, c.diameter)
c.radius = 5.0
print(c.radius, c.area)
try:
    c.radius = -1.0
except ValueError as e:
    print("error:", e)
try:
    c.area = 1.0
except AttributeError as e:
    print("AttributeError:", e)

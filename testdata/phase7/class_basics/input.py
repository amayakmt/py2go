class Point:
    count: int = 0

    def __init__(self, x: int, y: int) -> None:
        self.x = x
        self.y = y
        Point.count += 1

    def norm2(self) -> int:
        return self.x * self.x + self.y * self.y

    def __repr__(self) -> str:
        return f"Point({self.x}, {self.y})"

    def __add__(self, other: "Point") -> "Point":
        return Point(self.x + other.x, self.y + other.y)


p: Point = Point(3, 4)
q: Point = Point(1, 2)
print(p, q, p.norm2(), p + q)
print(Point.count)
p.x = 10
print(p.x, p)

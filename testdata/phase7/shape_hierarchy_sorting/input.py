class Shape:
    def area(self) -> float:
        raise NotImplementedError

    def perimeter(self) -> float:
        raise NotImplementedError

    def __lt__(self, other: "Shape") -> bool:
        return self.area() < other.area()

    def __repr__(self) -> str:
        return f"{type(self).__name__}(area={self.area():.2f})"


class Circle(Shape):
    def __init__(self, r: float) -> None:
        self.r = r

    def area(self) -> float:
        return 3.14159 * self.r ** 2

    def perimeter(self) -> float:
        return 2 * 3.14159 * self.r


class Rect(Shape):
    def __init__(self, w: float, h: float) -> None:
        self.w = w
        self.h = h

    def area(self) -> float:
        return self.w * self.h

    def perimeter(self) -> float:
        return 2 * (self.w + self.h)


class Square(Rect):
    def __init__(self, s: float) -> None:
        super().__init__(s, s)


shapes: list[Shape] = [Circle(1.0), Rect(2.0, 3.0), Square(1.5), Circle(0.5)]
print(sorted(shapes))
print([f"{s.perimeter():.2f}" for s in shapes])
print(max(shapes), min(shapes))

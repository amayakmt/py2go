class Shape:
    pass


class Square(Shape):
    def __init__(self, side: int) -> None:
        self.side = side


class Circle(Shape):
    def __init__(self, radius: int) -> None:
        self.radius = radius


def area(s: Shape) -> int:
    if isinstance(s, Square):
        return s.side * s.side
    if isinstance(s, Circle):
        return 3 * s.radius * s.radius
    return 0


shapes: list[Shape] = [Square(2), Circle(3), Shape()]
print([area(s) for s in shapes])
sq: Square = Square(4)
print(isinstance(sq, Shape), isinstance(sq, Square), isinstance(sq, Circle), isinstance(sq, (Circle, Square)))
print(type(sq).__name__, type(shapes[2]).__name__, sq.__class__.__name__)
print(type(sq) == Square, type(sq) is Shape, type(shapes[0]) == type(sq))
print(issubclass(Square, Shape), issubclass(Shape, Square), issubclass(Circle, Circle))

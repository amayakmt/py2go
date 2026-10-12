from dataclasses import dataclass


@dataclass
class Point:
    x: int
    y: int


@dataclass
class Circle:
    center: Point
    radius: int


class Square:
    __match_args__ = ("side",)

    def __init__(self, side: int) -> None:
        self.side = side


class Shape:
    pass


def area(s: object) -> str:
    match s:
        case Point(0, 0):
            return "origin"
        case Point(x=0, y=y):
            return "on y axis at " + str(y)
        case Point(x, y):
            return f"point {x},{y}"
        case Circle(Point(0, 0), r):
            return "centered circle " + str(r)
        case Circle(center=Point(x=cx), radius=r) if r > 5:
            return f"big circle at {cx}"
        case Circle():
            return "circle"
        case Square(side):
            return "square " + str(side * side)
        case int(n):
            return "int " + str(n)
        case str() as text:
            return "str " + text
        case _:
            return "unknown"


things: list[object] = [Point(0, 0), Point(0, 4), Point(2, 3), Circle(Point(0, 0), 2), Circle(Point(5, 1), 9), Circle(Point(1, 1), 1), Square(3), 42, "hi", 2.5, Shape()]
for obj in things:
    print(area(obj))

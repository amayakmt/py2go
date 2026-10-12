class Point:
    __match_args__ = ("x", "y")
    def __init__(self, x, y):
        self.x = x
        self.y = y

class Circle:
    __match_args__ = ("center", "r")
    def __init__(self, center, r):
        self.center = center
        self.r = r

def describe(shape):
    match shape:
        case Point(0, 0):
            return "origin"
        case Point(x, 0):
            return "on x axis at " + str(x)
        case Point(x=0, y=y):
            return "on y axis at " + str(y)
        case Point(x, y) if x == y:
            return "diagonal " + str(x)
        case Point():
            return "point"
        case Circle(Point(0, 0), r):
            return "circle at origin r=" + str(r)
        case Circle(center=Point(x, y), r=r):
            return f"circle {x},{y} r={r}"
        case [Point() as p, *_]:
            return "points starting with " + describe(p)
        case _:
            return "unknown"

for s in (Point(0, 0), Point(3, 0), Point(0, 4), Point(2, 2), Point(1, 5), Circle(Point(0, 0), 2), Circle(Point(1, 2), 3), [Point(0, 0), 1], "x"):
    print(describe(s))

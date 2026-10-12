from dataclasses import dataclass


@dataclass
class Point:
    x: int
    y: int = 0

    def dist2(self) -> int:
        return self.x * self.x + self.y * self.y


p: Point = Point(3, 4)
q: Point = Point(3)
print(p, q, p.dist2())
print(p == Point(3, 4), p == q, p != q)
p.x = 10
print(p, repr(p))

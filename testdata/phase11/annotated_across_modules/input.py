from models import Point, origin, SCALE
import models

def far(p: Point) -> bool:
    return p.dist2() > models.SCALE * SCALE

pts: list[Point] = [Point(1, 2), Point(3, 4), origin()]
print([far(p) for p in pts], [p.dist2() for p in pts], models.origin().x, SCALE)

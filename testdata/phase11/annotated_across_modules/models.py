class Point:
    def __init__(self, x: int, y: int) -> None:
        self.x = x
        self.y = y
    def dist2(self) -> int:
        return self.x * self.x + self.y * self.y

def origin() -> Point:
    return Point(0, 0)

SCALE: int = 3

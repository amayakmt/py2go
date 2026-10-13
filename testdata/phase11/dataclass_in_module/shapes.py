from dataclasses import dataclass

@dataclass
class Rect:
    w: int
    h: int
    def area(self) -> int:
        return self.w * self.h

def unit() -> Rect:
    return Rect(1, 1)

from dataclasses import dataclass


@dataclass(frozen=True)
class P:
    x: int
    y: int


pts: list[P] = [P(1, 2), P(3, 4), P(1, 2)]
print(P(1, 2) in pts, P(9, 9) in pts, pts.count(P(1, 2)), pts.index(P(3, 4)))
pts.remove(P(1, 2))
print(pts, len(pts))
print(sorted(pts, key=lambda p: (p.x, p.y)), max(pts, key=lambda p: p.y))
print(hash(P(1, 2)) == hash(P(1, 2)), hash(P(1, 2)) == hash(P(2, 1)))

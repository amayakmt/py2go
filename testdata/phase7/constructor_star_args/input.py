class P:
    def __init__(self, x: int, y: int = 5, *, z: int = 7) -> None:
        self.x = x
        self.y = y
        self.z = z

    def __repr__(self) -> str:
        return f"P({self.x}, {self.y}, {self.z})"


coords: list[int] = [1, 2]
opts: dict[str, int] = {"z": 9}
print(P(*coords), P(1), P(1, z=3), P(*coords, **opts), P(**{"x": 4}))
ps: list[P] = [P(i) for i in range(3)]
print(ps)
fallback: P = P(0)


def make(p: P = fallback, n: int = 2) -> str:
    return repr(p) * n


print(make(), make(P(1), 1))

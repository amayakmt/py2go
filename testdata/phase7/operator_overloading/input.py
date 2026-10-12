class Vec:
    def __init__(self, x: float, y: float) -> None:
        self.x = x
        self.y = y

    def __add__(self, other: "Vec") -> "Vec":
        return Vec(self.x + other.x, self.y + other.y)

    def __sub__(self, other: "Vec") -> "Vec":
        return Vec(self.x - other.x, self.y - other.y)

    def __mul__(self, k: float) -> "Vec":
        return Vec(self.x * k, self.y * k)

    def __rmul__(self, k: float) -> "Vec":
        return self * k

    def __neg__(self) -> "Vec":
        return Vec(-self.x, -self.y)

    def __abs__(self) -> float:
        return (self.x * self.x + self.y * self.y) ** 0.5

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Vec):
            return NotImplemented
        return self.x == other.x and self.y == other.y

    def __repr__(self) -> str:
        return f"Vec({self.x}, {self.y})"


a: Vec = Vec(1.0, 2.0)
b: Vec = Vec(3.0, 4.0)
print(a + b, a - b, a * 2.0, 3.0 * a, -a)
print(abs(b), a == Vec(1.0, 2.0), a == b, a != b, a == 5)
c: Vec = a
c += b
print(c, a)

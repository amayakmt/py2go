class Meters:
    def __init__(self, v: float) -> None:
        self.v = v

    def __add__(self, other: "Meters") -> "Meters":
        return Meters(self.v + other.v)

    def __radd__(self, other: int) -> "Meters":
        return Meters(self.v + other)

    def __mul__(self, k: float) -> "Meters":
        return Meters(self.v * k)

    def __rmul__(self, k: float) -> "Meters":
        return Meters(self.v * k)

    def __rsub__(self, other: float) -> "Meters":
        return Meters(other - self.v)

    def __repr__(self) -> str:
        return f"{self.v}m"


m: Meters = Meters(2.5)
print(m + m, 3 + m, m * 2.0, 2.0 * m, 10.0 - m)
total: Meters = Meters(0.0)
for x in [Meters(1.0), Meters(2.0), Meters(4.5)]:
    total = total + x
print(total)

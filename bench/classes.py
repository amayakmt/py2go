class Vec:
    def __init__(self, x: float, y: float):
        self.x = x
        self.y = y

    def add(self, o: "Vec") -> "Vec":
        return Vec(self.x + o.x, self.y + o.y)

    def norm2(self) -> float:
        return self.x * self.x + self.y * self.y


acc = Vec(0.0, 0.0)
step = Vec(0.5, 0.25)
for _ in range(2000000):
    acc = acc.add(step)
print(acc.norm2())

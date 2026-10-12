class Vec:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __add__(self, other):
        return Vec(self.x + other.x, self.y + other.y)

    def __sub__(self, other):
        return Vec(self.x - other.x, self.y - other.y)

    def __mul__(self, k):
        return Vec(self.x * k, self.y * k)

    def __rmul__(self, k):
        return self * k

    def __neg__(self):
        return Vec(-self.x, -self.y)

    def __eq__(self, other):
        return isinstance(other, Vec) and self.x == other.x and self.y == other.y

    def __abs__(self):
        return (self.x ** 2 + self.y ** 2) ** 0.5

    def __repr__(self):
        return f"Vec({self.x}, {self.y})"

def combine(a, b):
    return a + b, a - b, a * 2, 3 * b, -a, a == b, a != b

print(combine(Vec(1, 2), Vec(3, 4)))
print(abs(Vec(3, 4)), Vec(1, 1) == Vec(1, 1), Vec(1, 1) == "v", sum([Vec(1, 1), Vec(2, 2)], Vec(0, 0)))
v = Vec(1, 2)
w = v
v += Vec(1, 1)
print(v, w, v is w)

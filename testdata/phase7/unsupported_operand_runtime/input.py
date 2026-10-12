class V:
    def __init__(self, x: int) -> None:
        self.x = x

    def __add__(self, other: object) -> "V":
        if not isinstance(other, V):
            return NotImplemented
        return V(self.x + other.x)


a: V = V(1)
print((a + V(2)).x)
try:
    print((a + 1).x)
except TypeError as e:
    print("TypeError:", e)
try:
    print(1 + a)
except TypeError as e:
    print("TypeError:", e)

class Box:
    def __init__(self, v: int) -> None:
        self.v = v


a: Box = Box(1)
b: Box = Box(2)
try:
    print(a < b)
except TypeError as e:
    print("TypeError:", e)
try:
    print(sorted([b, a]))
except TypeError as e:
    print("TypeError:", e)
print(a == b, a == a, a != b)

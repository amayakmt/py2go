def to_f(c: float) -> float:
    return c * 9 / 5 + 32


for c in range(-40, 101, 35):
    print(c, "C =", to_f(c * 1.0), "F")

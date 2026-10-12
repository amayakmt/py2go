class N:
    def __init__(self, v: int) -> None:
        self.v = v

    def __add__(self, o: "N") -> "N":
        return N(self.v + o.v)

    def __sub__(self, o: "N") -> "N":
        return N(self.v - o.v)

    def __mul__(self, o: "N") -> "N":
        return N(self.v * o.v)

    def __truediv__(self, o: "N") -> float:
        return self.v / o.v

    def __floordiv__(self, o: "N") -> "N":
        return N(self.v // o.v)

    def __mod__(self, o: "N") -> "N":
        return N(self.v % o.v)

    def __pow__(self, o: "N") -> "N":
        return N(self.v ** 2)

    def __and__(self, o: "N") -> "N":
        return N(self.v & o.v)

    def __or__(self, o: "N") -> "N":
        return N(self.v | o.v)

    def __xor__(self, o: "N") -> "N":
        return N(self.v ^ o.v)

    def __lshift__(self, o: "N") -> "N":
        return N(self.v << o.v)

    def __rshift__(self, o: "N") -> "N":
        return N(self.v >> o.v)

    def __repr__(self) -> str:
        return "N" + str(self.v)


a: N = N(12)
b: N = N(5)
print(a + b, a - b, a * b, a / b, a // b, a % b)
print(a ** N(2), a & b, a | b, a ^ b, a << N(2), a >> N(2))

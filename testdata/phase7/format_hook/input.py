class Celsius:
    def __init__(self, t: float) -> None:
        self.t = t

    def __format__(self, spec: str) -> str:
        if spec == "f":
            return str(self.t * 9.0 / 5.0 + 32.0) + "F"
        return str(self.t) + "C"

    def __str__(self) -> str:
        return "T=" + str(self.t)


t: Celsius = Celsius(100.0)
print(f"{t}")
print(f"{t:f}")
print(f"{t:x}")
print(f"{t!s} {t!r}"[:6])

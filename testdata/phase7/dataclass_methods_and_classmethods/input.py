from dataclasses import dataclass


@dataclass
class Temp:
    deg: float
    unit: str = "C"

    @classmethod
    def boiling(cls) -> "Temp":
        return cls(100.0)

    @staticmethod
    def valid(unit: str) -> bool:
        return unit in ["C", "F"]

    def to_f(self) -> "Temp":
        return Temp(self.deg * 9.0 / 5.0 + 32.0, "F")

    def __add__(self, other: "Temp") -> "Temp":
        return Temp(self.deg + other.deg, self.unit)


b: Temp = Temp.boiling()
print(b, b.to_f(), Temp.valid("K"), Temp.valid("F"))
print(b + Temp(5.0))

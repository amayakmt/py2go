class Temperature:
    scale: str = "C"
    instances: int = 0

    def __init__(self, degrees: float) -> None:
        self.degrees = degrees
        Temperature.instances += 1

    @staticmethod
    def c_to_f(c: float) -> float:
        return c * 9.0 / 5.0 + 32.0

    @classmethod
    def freezing(cls) -> "Temperature":
        return cls(0.0)

    @classmethod
    def count(cls) -> int:
        return cls.instances

    def fahrenheit(self) -> float:
        return Temperature.c_to_f(self.degrees)


t: Temperature = Temperature.freezing()
u: Temperature = Temperature(100.0)
print(t.degrees, u.fahrenheit(), Temperature.c_to_f(37.0))
print(Temperature.count(), t.count(), u.scale)

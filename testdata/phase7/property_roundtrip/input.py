class Temperature:
    def __init__(self, c: float) -> None:
        self._c = c

    @property
    def celsius(self) -> float:
        return self._c

    @celsius.setter
    def celsius(self, v: float) -> None:
        self._c = v

    @property
    def fahrenheit(self) -> float:
        return self._c * 9.0 / 5.0 + 32.0

    @fahrenheit.setter
    def fahrenheit(self, f: float) -> None:
        self._c = (f - 32.0) * 5.0 / 9.0


t: Temperature = Temperature(100.0)
print(f"{t.celsius:.1f} {t.fahrenheit:.1f}")
t.fahrenheit = 32.0
print(f"{t.celsius:.1f}", t.fahrenheit)
t.celsius += 10.0
print(t.celsius, t.fahrenheit)

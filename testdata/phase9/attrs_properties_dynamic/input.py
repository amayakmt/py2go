class Temperature:
    def __init__(self, celsius):
        self._c = celsius

    @property
    def celsius(self):
        return self._c

    @celsius.setter
    def celsius(self, value):
        if value < -273.15:
            raise ValueError("below absolute zero")
        self._c = value

    @property
    def fahrenheit(self):
        return self._c * 9 / 5 + 32

t = Temperature(100)
print(t.celsius, t.fahrenheit)
t.celsius = 0
print(t.celsius, t.fahrenheit)
for bad in (-300, "cold"):
    try:
        t.celsius = bad
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
try:
    t.fahrenheit = 5
except AttributeError as e:
    print("AttributeError:", e)
print(getattr(t, "celsius"), hasattr(t, "fahrenheit"))
setattr(t, "celsius", 25)
print(t.celsius)

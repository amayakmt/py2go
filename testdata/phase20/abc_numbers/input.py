import numbers
from abc import ABC, ABCMeta, abstractmethod
from fractions import Fraction
from decimal import Decimal


class Shape(ABC):
    @abstractmethod
    def area(self):
        ...

    @abstractmethod
    def name(self):
        ...

    def describe(self):
        return "%s area=%s" % (self.name(), self.area())


class Sq(Shape):
    def __init__(self, s):
        self.s = s

    def area(self):
        return self.s * self.s

    def name(self):
        return "square"


class Half(Shape):
    def area(self):
        return 0


print(Sq(3).describe())
for cls in (Shape, Half):
    try:
        cls()
    except TypeError as e:
        print("TypeError", e)


class Base(metaclass=ABCMeta):
    @abstractmethod
    def go(self):
        ...


class Impl(Base):
    def go(self):
        return "went"


print(Impl().go())
try:
    Base()
except TypeError as e:
    print("TypeError", e)


class Duck:
    def area(self):
        return 1


Shape.register(Duck)
print(isinstance(Duck(), Shape), issubclass(Duck, Shape), isinstance(3, Shape))

for v in (3, 2.5, 1j, True, "x", Fraction(1, 2), Decimal("1.5")):
    print(repr(v), isinstance(v, numbers.Number), isinstance(v, numbers.Integral), isinstance(v, numbers.Real), isinstance(v, numbers.Complex))
print(issubclass(int, numbers.Integral), issubclass(float, numbers.Rational), issubclass(bool, numbers.Number))


def kind(x):
    if isinstance(x, numbers.Integral):
        return "integral"
    if isinstance(x, numbers.Real):
        return "real"
    return "other"


print([kind(v) for v in (1, 1.5, "s")])

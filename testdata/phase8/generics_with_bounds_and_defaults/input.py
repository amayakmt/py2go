class Animal:
    def __init__(self, n: str) -> None:
        self.n = n

    def speak(self) -> str:
        return self.n + " speaks"


class Dog(Animal):
    pass


class Shelter[A: Animal]:
    def __init__(self) -> None:
        self.animals: list[A] = []

    def admit(self, a: A) -> None:
        self.animals.append(a)

    def chorus(self) -> list[str]:
        return [a.speak() for a in self.animals]


class Cell[T = int]:
    def __init__(self, v: T) -> None:
        self.v = v


s: Shelter[Dog] = Shelter()
s.admit(Dog("rex"))
s.admit(Dog("fido"))
print(s.chorus())
c: Cell[int] = Cell(5)
print(c.v)


def twice[T: (int, str)](x: T) -> T:
    return x + x


print(twice(4), twice("ab"))

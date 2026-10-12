class Dog:
    def __init__(self, name: str) -> None:
        self.name = name

    def bark(self) -> str:
        return self.name + " barks"


class Cat:
    def __init__(self, lives: int) -> None:
        self.lives = lives


def describe(x: object) -> str:
    if isinstance(x, int):
        return "int " + str(x + 1)
    if isinstance(x, str):
        return "str " + x.upper()
    if isinstance(x, Dog):
        return x.bark()
    if isinstance(x, Cat):
        return "cat with " + str(x.lives)
    return "something else"


print(describe(41))
print(describe("hi"))
print(describe(Dog("rex")))
print(describe(Cat(9)))
print(describe(2.5))
print(describe([1, 2]))
print(describe(True))


def first_dog(x: object) -> str:
    if not isinstance(x, Dog):
        return "no dog"
    return x.name


print(first_dog(Dog("fido")), first_dog(3))

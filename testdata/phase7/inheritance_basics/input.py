class Animal:
    def __init__(self, name: str) -> None:
        self.name = name

    def speak(self) -> str:
        return "..."

    def describe(self) -> str:
        return self.name + " says " + self.speak()


class Dog(Animal):
    def speak(self) -> str:
        return "woof"


class Cat(Animal):
    def __init__(self, name: str, lives: int) -> None:
        super().__init__(name)
        self.lives = lives

    def speak(self) -> str:
        return "meow"


animals: list[Animal] = [Animal("generic"), Dog("rex"), Cat("tom", 9)]
for a in animals:
    print(a.describe())
cat: Cat = Cat("kit", 7)
print(cat.lives, cat.name, cat.describe())

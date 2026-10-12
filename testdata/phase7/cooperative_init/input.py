class Named:
    def __init__(self, name: str) -> None:
        self.name = name


class Aged:
    def __init__(self, age: int) -> None:
        self.age = age


class Person(Named, Aged):
    def __init__(self, name: str, age: int) -> None:
        Named.__init__(self, name)
        Aged.__init__(self, age)

    def card(self) -> str:
        return self.name + "/" + str(self.age)


p: Person = Person("ann", 30)
print(p.card())

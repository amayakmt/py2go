from typing import Optional


class Animal:
    def __init__(self, name: str) -> None:
        self.name = name

    def sound(self) -> str:
        return "..."


class Dog(Animal):
    def sound(self) -> str:
        return "woof"


def adopt(name: str, kind: str) -> Optional[Animal]:
    if kind == "dog":
        return Dog(name)
    if kind == "any":
        return Animal(name)
    return None


for k in ["dog", "any", "fish"]:
    pet: Optional[Animal] = adopt("rex", k)
    if pet is None:
        print(k, "-> nothing")
    else:
        print(k, "->", pet.name, pet.sound(), isinstance(pet, Dog))
pets: list[Optional[Animal]] = [adopt("a", "dog"), None, adopt("b", "any")]
print([p.sound() if p is not None else "-" for p in pets])

from base import Animal
class Dog(Animal):
    sound = "woof"
    def fetch(self):
        return self.name + " fetches"
class Cat(Animal):
    sound = "meow"

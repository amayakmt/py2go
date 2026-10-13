import base
import pets
from pets import Dog

class Puppy(Dog):
    def speak(self):
        return super().speak() + "!"

animals = [Dog("rex"), pets.Cat("tom"), Puppy("bit"), base.Animal("x")]
for a in animals:
    print(a.speak(), a, isinstance(a, base.Animal), isinstance(a, Dog))
print(animals[0].fetch(), Puppy.sound, [type(a).__name__ for a in animals])

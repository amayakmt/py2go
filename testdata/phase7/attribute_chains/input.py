class Engine:
    def __init__(self, hp: int) -> None:
        self.hp = hp


class Car:
    def __init__(self, name: str, hp: int) -> None:
        self.name = name
        self.engine = Engine(hp)


class Garage:
    def __init__(self) -> None:
        self.cars: list[Car] = []
        self.by_name: dict[str, Car] = {}

    def add(self, c: Car) -> None:
        self.cars.append(c)
        self.by_name[c.name] = c


g: Garage = Garage()
g.add(Car("a", 100))
g.add(Car("b", 200))
g.cars[0].engine.hp += 50
g.by_name["b"].engine.hp = g.by_name["b"].engine.hp * 2
print(g.cars[0].engine.hp, g.cars[1].engine.hp, g.by_name["a"].engine.hp)
print(sum(c.engine.hp for c in g.cars), len(g.cars), g.cars[-1].name)

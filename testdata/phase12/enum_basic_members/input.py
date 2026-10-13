from enum import Enum, auto, unique
class Color(Enum):
    RED = 1
    GREEN = 2
    BLUE = 3
    def describe(self):
        return f"{self.name}:{self.value}"
    @property
    def lower(self):
        return self.name.lower()
    @classmethod
    def default(cls):
        return cls.GREEN
class Planet(Enum):
    EARTH = (5.97e24, 6.37e6)
    MARS = (6.42e23, 3.39e6)
    def __init__(self, mass, radius):
        self.mass = mass
        self.radius = radius
    def gravity(self):
        return round(6.674e-11 * self.mass / (self.radius * self.radius), 2)
class Dir(Enum):
    N = "north"
    S = "south"
    OTHER = 1
    ALIAS = "north"
print(Color.RED, repr(Color.RED), Color.RED.name, Color.RED.value, Color(2), Color["BLUE"], Color.RED is Color.RED, Color.RED == Color.GREEN, Color.RED != Color.GREEN, hash(Color.RED) == hash(Color.RED), type(Color.RED).__name__, isinstance(Color.RED, Color), isinstance(Color.RED, Enum))
print(list(Color), [c.name for c in Color], len(Color), Color.RED in Color if False else "x", [c.describe() for c in Color], Color.GREEN.lower, Color.default(), {Color.RED: "r"}[Color.RED], sorted(Color, key=lambda c: -c.value))
print(Planet.EARTH.mass, Planet.MARS.radius, Planet.EARTH.gravity(), Planet.MARS.gravity(), list(Planet), Planet.EARTH.name)
print(Dir.N.value, Dir.OTHER.value, Dir.ALIAS is Dir.N, Dir("south"), list(Dir), [d.name for d in Dir], list(Dir.__members__), Dir.__members__["ALIAS"] is Dir.N)
for bad in (lambda: Color(9), lambda: Color["PURPLE"], lambda: Dir("up")):
    try:
        print(bad())
    except (ValueError, KeyError) as e:
        print(type(e).__name__, e)
def paint(c):
    if c is Color.RED:
        return "hot"
    elif c == Color.BLUE:
        return "cold"
    return "meh"
print([paint(c) for c in Color], Color.RED.describe(), f"{Color.RED}", f"{Color.RED!r}", str(Dir.S), "%s" % Color.BLUE)
match_result = {Color.RED: 1, Color.GREEN: 2}
print(sorted(match_result.items(), key=lambda kv: kv[0].value), Color.RED in match_result, Color.BLUE in match_result)

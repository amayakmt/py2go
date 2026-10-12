from dataclasses import dataclass, field


@dataclass(frozen=True)
class Pair:
    a: int
    b: str


@dataclass(order=True)
class Card:
    rank: int
    suit: str = "s"


@dataclass
class Bag:
    name: str
    items: list[int] = field(default_factory=list)
    tag: str = field(default="t", repr=False)


p: Pair = Pair(1, "x")
print(p, hash(p) == hash(Pair(1, "x")), p == Pair(1, "x"))
try:
    p.a = 5
except AttributeError as e:
    print(type(e).__name__, e)
cards: list[Card] = [Card(3), Card(1, "h"), Card(2)]
print(sorted(cards), Card(1) < Card(2), max(cards))
b1: Bag = Bag("one")
b2: Bag = Bag("two")
b1.items.append(7)
print(b1, b2, b1.items is b2.items)

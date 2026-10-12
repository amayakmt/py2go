from dataclasses import dataclass, field


@dataclass
class Item:
    name: str
    qty: int = 1
    price: float = 0.0


@dataclass
class Order:
    id: int
    items: list[Item] = field(default_factory=list)
    meta: dict[str, int] = field(default_factory=dict)

    def total(self) -> float:
        return sum(i.qty * i.price for i in self.items)


o: Order = Order(1)
o.items.append(Item("pen", 3, 1.5))
o.items.append(Item("book", price=12.0))
o.meta["rush"] = 1
print(o)
print(o.total(), len(o.items), o.items[0] == Item("pen", 3, 1.5))
orders: list[Order] = [o, Order(2, [Item("x")])]
print(orders)
print(sorted(orders, key=lambda x: x.id, reverse=True)[0].id)

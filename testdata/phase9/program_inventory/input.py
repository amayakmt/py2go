class Item:
    def __init__(self, name, price, qty=0, tags=None):
        self.name = name
        self.price = price
        self.qty = qty
        self.tags = tags or []

    def total(self):
        return self.price * self.qty

    def __repr__(self):
        return "Item(%r, %s x%d)" % (self.name, self.price, self.qty)

class Inventory:
    def __init__(self):
        self.items = {}
        self.log = []

    def add(self, item):
        if item.name in self.items:
            self.items[item.name].qty += item.qty
            self.log.append(("restock", item.name, item.qty))
        else:
            self.items[item.name] = item
            self.log.append(("new", item.name, item.qty))

    def remove(self, name, qty):
        item = self.items.get(name)
        if item is None:
            raise KeyError(name)
        if item.qty < qty:
            raise ValueError("only %d of %s" % (item.qty, name))
        item.qty -= qty
        self.log.append(("sell", name, qty))
        return item.price * qty

    def value(self):
        return sum(i.total() for i in self.items.values())

    def with_tag(self, tag):
        return sorted(i.name for i in self.items.values() if tag in i.tags)

inv = Inventory()
inv.add(Item("bolt", 0.25, 100, ["metal", "small"]))
inv.add(Item("nut", 0.1, 250, ["metal", "small"]))
inv.add(Item("plank", 12.5, 8, ["wood"]))
inv.add(Item("bolt", 0.25, 50))
print(inv.value(), inv.with_tag("metal"), inv.with_tag("wood"), inv.with_tag("none"))
print(inv.remove("plank", 3), inv.items["plank"])
for name, qty in (("bolt", 500), ("gear", 1), ("nut", 50)):
    try:
        print("sold", inv.remove(name, qty))
    except (KeyError, ValueError) as e:
        print("error", type(e).__name__, e)
print(inv.log)
print(sorted(inv.items.values(), key=lambda i: -i.total())[0])

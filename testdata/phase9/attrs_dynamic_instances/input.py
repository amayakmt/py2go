class Bag:
    pass

b = Bag()
b.x = 1
b.name = "bag"
b.items = [1, 2]
print(b.x, b.name, b.items, hasattr(b, "x"), hasattr(b, "zzz"))
setattr(b, "dyn", 42)
print(getattr(b, "dyn"), getattr(b, "none", "default"), b.dyn)
b.x += 10
b.items.append(3)
print(b.x, b.items)
delattr(b, "dyn")
print(hasattr(b, "dyn"))
del b.x
print(hasattr(b, "x"))
try:
    print(b.x)
except AttributeError as e:
    print("AttributeError:", e)
other = Bag()
other.x = "different"
print(other.x, hasattr(other, "name"))
for attr in ("name", "items", "q"):
    print(attr, getattr(b, attr, "?"))

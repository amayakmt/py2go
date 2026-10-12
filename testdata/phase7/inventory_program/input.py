class OutOfStock(Exception):
    def __init__(self, sku: str, wanted: int, have: int) -> None:
        super().__init__(f"{sku}: wanted {wanted}, have {have}")
        self.sku = sku


class Product:
    def __init__(self, sku: str, name: str, price: float, qty: int) -> None:
        self.sku = sku
        self.name = name
        self.price = price
        self.qty = qty

    def value(self) -> float:
        return self.price * self.qty


class Inventory:
    def __init__(self) -> None:
        self.products: dict[str, Product] = {}

    def add(self, p: Product) -> None:
        if p.sku in self.products:
            self.products[p.sku].qty += p.qty
        else:
            self.products[p.sku] = p

    def remove(self, sku: str, n: int) -> None:
        p: Product = self.products[sku]
        if p.qty < n:
            raise OutOfStock(sku, n, p.qty)
        p.qty -= n

    def report(self) -> list[str]:
        return [f"{p.sku} {p.name} x{p.qty} = {p.value():.2f}" for p in sorted(self.products.values(), key=lambda p: p.sku)]


inv: Inventory = Inventory()
inv.add(Product("b1", "bolt", 0.25, 100))
inv.add(Product("a1", "anvil", 99.5, 2))
inv.add(Product("b1", "bolt", 0.25, 50))
inv.remove("b1", 120)
for line in inv.report():
    print(line)
try:
    inv.remove("a1", 3)
except OutOfStock as e:
    print(e, e.sku)
try:
    inv.remove("zz", 1)
except KeyError as e:
    print("KeyError", e)

class Discount:
    def apply(self, price: float) -> float:
        return price


class Percent(Discount):
    def __init__(self, pct: float) -> None:
        self.pct = pct

    def apply(self, price: float) -> float:
        return price * (1.0 - self.pct / 100.0)


class Flat(Discount):
    def __init__(self, amount: float) -> None:
        self.amount = amount

    def apply(self, price: float) -> float:
        return max(price - self.amount, 0.0)


class Cart:
    def __init__(self) -> None:
        self.prices: list[float] = []
        self.discount: Discount = Discount()

    def total(self) -> float:
        return self.discount.apply(sum(self.prices))


cart: Cart = Cart()
cart.prices = [10.0, 20.0, 12.5]
print(cart.total())
cart.discount = Percent(10.0)
print(cart.total())
cart.discount = Flat(100.0)
print(cart.total())

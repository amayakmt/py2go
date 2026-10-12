class Bag:
    def __init__(self, n: int) -> None:
        self.n = n

    def __len__(self) -> int:
        return self.n


empty: Bag = Bag(0)
full: Bag = Bag(3)
print(bool(empty), bool(full), not empty, not full)
if empty:
    print("empty is truthy")
else:
    print("empty is falsy")
count: int = 0
while full:
    full.n -= 1
    count += 1
print(count, len(full), len(Bag(7)))
plain: object = Bag(0)
class P:
    pass
print(bool(P()))

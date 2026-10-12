class Key:
    def __init__(self, a: int, b: int) -> None:
        self.a = a
        self.b = b

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Key):
            return NotImplemented
        return self.a == other.a and self.b == other.b

    def __hash__(self) -> int:
        return hash((self.a, self.b))

    def __repr__(self) -> str:
        return f"Key({self.a}, {self.b})"


k1: Key = Key(1, 2)
k2: Key = Key(1, 2)
k3: Key = Key(2, 1)
print(k1 == k2, k1 is k2, k1 != k3, hash(k1) == hash(k2))
keys: list[Key] = [k1, k3]
print(k2 in keys, Key(9, 9) in keys, keys.index(k2), keys.count(k2))
print(keys == [k2, k3], keys == [k3, k2])

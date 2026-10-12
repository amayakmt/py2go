class Plain:
    pass


class Hashed:
    def __init__(self, n: int) -> None:
        self.n = n

    def __hash__(self) -> int:
        return self.n * 7

    def __eq__(self, other: object) -> bool:
        return isinstance(other, Hashed) and other.n == self.n


class Unhashable:
    def __eq__(self, other: object) -> bool:
        return True


print(hash(Hashed(3)), hash(Hashed(3)) == hash(Hashed(3)))
p: Plain = Plain()
print(hash(p) == hash(p))
print(hash(1), hash("a") == hash("a"), hash((1, 2)) == hash((1, 2)))
seen: dict[Plain, int] = {}
seen[p] = 1
q: Plain = Plain()
seen[q] = 2
print(len(seen), seen[p], seen[q])
print(p in seen, Plain() in seen)

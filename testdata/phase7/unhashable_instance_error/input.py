class Unhashable:
    def __eq__(self, other: object) -> bool:
        return True


u: Unhashable = Unhashable()
print(u == u)
print(hash(u))

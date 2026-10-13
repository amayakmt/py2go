class K:
    def __eq__(self, other: object) -> bool:
        return True

    def __hash__(self) -> int:
        return 1


d: dict[K, int] = {}
d[K()] = 1
d[K()] = 2
print(len(d), d[K()])

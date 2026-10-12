class Level:
    def __init__(self, n: int) -> None:
        self.n = n

    def __lt__(self, o: "Level") -> bool:
        return self.n < o.n

    def __gt__(self, o: "Level") -> bool:
        return self.n > o.n

    def __ge__(self, o: "Level") -> bool:
        return self.n >= o.n


class Other:
    def __lt__(self, o: "Level") -> bool:
        return True

    def __gt__(self, o: "Level") -> bool:
        return False


lo: Level = Level(1)
hi: Level = Level(5)
print(lo < hi, lo > hi, hi >= lo, hi >= hi)
try:
    print(lo <= hi)
except TypeError as e:
    print("TypeError:", e)
o: Other = Other()
print(o < lo, o > lo)

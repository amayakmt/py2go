class Stats:
    def __init__(self) -> None:
        self.n = 0
        self.total = 0.0
        self.text = ""
        self.counts: dict[str, int] = {}
        self.hist: list[int] = [0, 0, 0]

    def add(self, key: str, x: float) -> None:
        self.n += 1
        self.total += x
        self.text += key
        if key not in self.counts:
            self.counts[key] = 0
        self.counts[key] += 1
        self.hist[self.n % 3] += 1
        self.hist[0] -= 1
        self.total *= 1.0
        self.n <<= 0


s: Stats = Stats()
for k, x in [("a", 1.5), ("b", 2.5), ("a", 3.0), ("c", 0.5)]:
    s.add(k, x)
print(s.n, s.total, s.text, s.counts, s.hist)

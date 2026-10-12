class Range3Iter:
    def __init__(self, n: int) -> None:
        self.i = 0
        self.n = n

    def __next__(self) -> int:
        if self.i >= self.n:
            raise StopIteration
        self.i += 1
        return self.i * 3


class Range3:
    def __init__(self, n: int) -> None:
        self.n = n

    def __iter__(self) -> Range3Iter:
        return Range3Iter(self.n)


r: Range3 = Range3(4)
print(list(r), list(r))
for a in r:
    for b in r:
        if a == b:
            print(a)

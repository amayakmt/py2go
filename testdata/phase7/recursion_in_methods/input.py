class Walker:
    def __init__(self) -> None:
        self.depth = 0

    def down(self, n: int) -> int:
        self.depth += 1
        return self.down(n + 1)

    def fact(self, n: int) -> int:
        if n <= 1:
            return 1
        return n * self.fact(n - 1)


w: Walker = Walker()
print(w.fact(10))
try:
    w.down(0)
except RecursionError as e:
    print("RecursionError:", e, w.depth > 100)

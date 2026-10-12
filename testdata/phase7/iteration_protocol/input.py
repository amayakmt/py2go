class Countdown:
    def __init__(self, start: int) -> None:
        self.current = start

    def __iter__(self) -> "Countdown":
        return self

    def __next__(self) -> int:
        if self.current <= 0:
            raise StopIteration
        self.current -= 1
        return self.current + 1


for n in Countdown(3):
    print(n)
print(list(Countdown(4)), sum(Countdown(4)), max(Countdown(5)))
print([x * x for x in Countdown(3)], sorted(Countdown(3)))
total: int = 0
for v in Countdown(10):
    if v % 2 == 0:
        continue
    total += v
print(total)

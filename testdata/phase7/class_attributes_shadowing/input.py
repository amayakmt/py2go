class Counter:
    total: int = 0
    label = "counter"

    def __init__(self, start: int) -> None:
        self.value = start
        Counter.total += start

    def bump(self) -> int:
        self.value += 1
        Counter.total += 1
        return self.value

    def rename(self, text: str) -> None:
        self.label = text


a: Counter = Counter(5)
b: Counter = Counter(10)
print(Counter.total, Counter.label, a.label, b.label)
a.rename("a-only")
print(Counter.label, a.label, b.label)
print(a.bump(), b.bump(), Counter.total)
Counter.label = "changed"
print(Counter.label, a.label, b.label)

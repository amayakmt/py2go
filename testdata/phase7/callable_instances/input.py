class Multiplier:
    def __init__(self, k: int) -> None:
        self.k = k
        self.calls = 0

    def __call__(self, x: int) -> int:
        self.calls += 1
        return x * self.k


double: Multiplier = Multiplier(2)
print(double(4), double(10), double.calls)
values: list[int] = [double(i) for i in range(4)]
print(values, double.calls, callable(double))

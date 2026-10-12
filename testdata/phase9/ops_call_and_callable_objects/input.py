class Adder:
    def __init__(self, n):
        self.n = n

    def __call__(self, x, scale=1):
        return (x + self.n) * scale

add5 = Adder(5)
print(add5(1), add5(1, scale=2), add5(x=0), callable(add5))
print(list(map(add5, [1, 2, 3])), sorted([3, 1, 2], key=add5), [f(10) for f in (add5, Adder(1), len)] if False else "skip")
funcs = {"five": add5, "one": Adder(1), "id": lambda x: x}
for name in sorted(funcs):
    print(name, funcs[name](10))

def compose(*fs):
    def run(x):
        for f in reversed(fs):
            x = f(x)
        return x
    return run

inc, dbl = Adder(1), lambda x: x * 2
print(compose(inc, dbl)(5), compose(dbl, inc)(5), compose()(7))
class Counter:
    def __init__(self):
        self.calls = 0
    def __call__(self):
        self.calls += 1
        return self.calls
c = Counter()
print(c(), c(), c(), c.calls)

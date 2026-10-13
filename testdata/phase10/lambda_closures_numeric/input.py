adders = [lambda x, k=k: x + k for k in (1, 2 ** 70, 0.5)]
print([f(10) for f in adders], [f(2 ** 70) for f in adders], [f(True) for f in adders])
def make_counter(step):
    count = 0
    def inc():
        nonlocal count
        count += step
        return count
    return inc
c1, c2 = make_counter(1), make_counter(2 ** 40)
print([c1() for _ in range(3)], [c2() for _ in range(3)], c2() > 2 ** 41)
def compose2(f, g):
    return lambda x: g(f(x))
sq, inc, neg = (lambda x: x * x), (lambda x: x + 1), (lambda x: -x)
pipeline = compose2(compose2(sq, inc), neg)
print([pipeline(x) for x in (0, 1, 2, 2 ** 40, 1.5, True)], list(map(compose2(inc, sq), range(5))), sorted([3, -1, 2 ** 70, -2 ** 70], key=lambda v: -abs(v)), list(filter(lambda v: v & 1, [2 ** 70 + i for i in range(6)])))
total = 0
for f in (sq, inc, neg, abs, float, int, str, repr, bool, hex, bin):
    try:
        print(f(12), f(-3), end=" | ")
    except TypeError as e:
        print("T", end=" | ")
print()
print((lambda: 2 ** 100)(), (lambda x: x if x else None)(0), (lambda x, y: x ** y)(2, 100))

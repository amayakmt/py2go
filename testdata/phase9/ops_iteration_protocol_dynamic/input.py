class Countdown:
    def __init__(self, n):
        self.n = n

    def __iter__(self):
        return self

    def __next__(self):
        if self.n <= 0:
            raise StopIteration
        self.n -= 1
        return self.n + 1

print(list(Countdown(3)), sum(Countdown(4)), max(Countdown(5)))
it = Countdown(2)
print(next(it), next(it), next(it, "done"))

class Evens:
    def __init__(self, limit):
        self.limit = limit
    def __iter__(self):
        n = 0
        while n < self.limit:
            yield n
            n += 2

e = Evens(7)
print(list(e), list(e), [x * x for x in e], dict(zip("abcd", e)))
print(sorted(Evens(5), reverse=True), tuple(Evens(5)), 4 in Evens(9), set(Evens(5)) == {0, 2, 4})
a, *rest = Evens(9)
print(a, rest)
for i, x in enumerate(Countdown(3), 10):
    print(i, x)

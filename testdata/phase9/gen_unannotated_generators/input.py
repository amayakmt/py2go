def count(start, stop=None):
    n = start
    while stop is None or n < stop:
        yield n
        n += 1

def take(it, n):
    out = []
    for x in it:
        if len(out) == n:
            break
        out.append(x)
    return out

print(take(count(5), 3), list(count(2, 5)), sum(count(1, 101)))

def fib():
    a, b = 0, 1
    while True:
        yield a
        a, b = b, a + b

print(take(fib(), 10))
g = fib()
print(next(g), next(g), next(g), [next(g) for _ in range(5)])

def echo():
    received = []
    while True:
        x = yield len(received)
        if x is None:
            return received
        received.append(x)

e = echo()
print(next(e), e.send("a"), e.send("b"))
try:
    e.send(None)
except StopIteration as stop:
    print("stopped with", stop.value)

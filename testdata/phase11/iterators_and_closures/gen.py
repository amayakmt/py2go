class Counter:
    def __init__(self, n):
        self.n = n
    def __iter__(self):
        self.i = 0
        return self
    def __next__(self):
        if self.i >= self.n:
            raise StopIteration
        self.i += 1
        return self.i

def fib():
    a, b = 0, 1
    while True:
        yield a
        a, b = b, a + b

def take(it, k):
    out = []
    for x in it:
        if len(out) >= k:
            break
        out.append(x)
    return out

def make_counter():
    c = 0
    def inc():
        nonlocal c
        c += 1
        return c
    return inc

import random
random.seed(12345)
print([random.random() for _ in range(3)], [random.randint(1, 6) for _ in range(10)], [random.randrange(100) for _ in range(5)], [random.randrange(10, 50, 5) for _ in range(5)], [random.randrange(-5, 5) for _ in range(5)])
print(random.uniform(0, 10), random.uniform(-1, 1), random.getrandbits(1), random.getrandbits(8), random.getrandbits(32), random.getrandbits(33), random.getrandbits(64), random.getrandbits(200), random.getrandbits(0))
random.seed(0)
print(random.random(), random.random(), random.randint(0, 10 ** 30), random.choice([10, 20, 30]), random.choice("xyz"), random.choice((1, 2, 3, 4)))
for seed in (1, -1, 2 ** 32, 2 ** 64 + 5, "text", "", b"bytes", bytearray(b"ba"), 1.5, -0.0, 10 ** 50):
    random.seed(seed)
    print(repr(seed), random.random(), random.randrange(1000))
a = random.Random(99)
b = random.Random(99)
print(a.random() == b.random(), a.randint(1, 10 ** 12) == b.randint(1, 10 ** 12), [a.getrandbits(5) for _ in range(5)] == [b.getrandbits(5) for _ in range(5)])
st = a.getstate()
x = [a.random() for _ in range(3)]
a.setstate(st)
print(x == [a.random() for _ in range(3)], len(st), st[0], len(st[1]), st[2])
c = random.Random()
d = random.Random()
print(c.random() != d.random() or True, 0 <= c.random() < 1, random.Random(5).randbytes(6), random.Random(5).randbytes(0))
for bad in (lambda: random.randrange(0), lambda: random.randrange(5, 5), lambda: random.randrange(1, 10, 0), lambda: random.randrange(10, 1), lambda: random.randrange(1.5), lambda: random.choice([]), lambda: random.getrandbits(-1), lambda: random.seed([1]), lambda: random.randint(5, 1), lambda: random.sample([1, 2], 3), lambda: random.sample({1, 2}, 1), lambda: random.randrange(1, 10, -1)):
    try:
        print(bad())
    except (ValueError, TypeError, IndexError) as e:
        print(type(e).__name__, e)

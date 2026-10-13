import gen
from gen import Counter, take, fib
print(list(Counter(3)), take(fib(), 10), take(gen.fib(), 3), sum(gen.Counter(4)))
c = gen.make_counter()
print(c(), c(), gen.make_counter()())

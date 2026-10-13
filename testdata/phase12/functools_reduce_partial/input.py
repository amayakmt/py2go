import functools, operator
from functools import reduce, partial, cmp_to_key, cache, lru_cache, wraps
print(reduce(operator.add, [1, 2, 3, 4]), reduce(operator.mul, range(1, 8)), reduce(lambda a, b: a + b, ["a", "b", "c"]), reduce(max, [3, 9, 2], 100), reduce(lambda a, b: a * 2 + b, [1, 0, 1, 1], 0), reduce(operator.add, [], 5), reduce(operator.add, [7]))
try:
    reduce(operator.add, [])
except TypeError as e:
    print(e)
p = partial(pow, 2)
q = partial(int, base=2)
r = partial(sorted, reverse=True)
print(p(10), p(0), q("1011"), r([3, 1, 2]), partial(max, 5)(1, 9), partial(print, "x", sep="-")("y") if False else "ok")
words = ["pear", "fig", "banana", "kiwi"]
print(sorted(words, key=cmp_to_key(lambda a, b: len(a) - len(b))), sorted(words, key=cmp_to_key(lambda a, b: (a > b) - (a < b))), max(words, key=cmp_to_key(lambda a, b: len(a) - len(b))))
@cache
def fib(n):
    return n if n < 2 else fib(n - 1) + fib(n - 2)
@lru_cache(maxsize=None)
def tri(n):
    return n if n == 0 else n + tri(n - 1)
@lru_cache
def sq(n):
    return n * n
print(fib(90), tri(200), sq(7), [fib(i) for i in range(10)])
def deco(f):
    @wraps(f)
    def wrapper(*a):
        return f(*a) * 2
    return wrapper
@deco
def three():
    return 3
print(three(), functools.reduce(operator.xor, [1, 2, 3]))

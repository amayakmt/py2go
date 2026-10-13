import itertools
from itertools import islice, chain, count, cycle, repeat, tee, groupby, accumulate, product, zip_longest
import operator
def take(n, iterable):
    return list(islice(iterable, n))
def prepend(value, iterator):
    return chain([value], iterator)
def tabulate(function, start=0):
    return map(function, count(start))
def nth(iterable, n, default=None):
    return next(islice(iterable, n, None), default)
def flatten(list_of_lists):
    return chain.from_iterable(list_of_lists)
def pad_none(iterable):
    return chain(iterable, repeat(None))
def ncycles(iterable, n):
    return chain.from_iterable(repeat(tuple(iterable), n))
def grouper(iterable, n, fillvalue=None):
    args = [iter(iterable)] * n
    return zip_longest(*args, fillvalue=fillvalue)
def roundrobin(*iterables):
    pending = len(iterables)
    nexts = cycle(iter(it).__next__ for it in iterables) if False else None
    out = []
    its = [iter(i) for i in iterables]
    while its:
        for it in list(its):
            try:
                out.append(next(it))
            except StopIteration:
                its.remove(it)
    return out
def powerset(iterable):
    s = list(iterable)
    return chain.from_iterable(itertools.combinations(s, r) for r in range(len(s) + 1))
def unique_everseen(iterable):
    seen = set()
    for x in iterable:
        if x not in seen:
            seen.add(x)
            yield x
def partition(pred, iterable):
    t1, t2 = tee(iterable)
    return list(itertools.filterfalse(pred, t1)), list(filter(pred, t2))
print(take(5, tabulate(lambda x: x * x)), take(3, prepend(0, count(10))), nth(count(), 7), nth("abc", 5, "none"), list(flatten([[1], [2, 3], []])), take(5, pad_none([1, 2])), list(ncycles("ab", 3)))
print(list(grouper("ABCDEFG", 3, "x")), roundrobin("ABC", "D", "EF"), list(powerset([1, 2, 3])), list(unique_everseen("AAAABBBCCDAABBB")), partition(lambda x: x % 2, range(10)))
print(list(accumulate(range(1, 8), operator.mul)), [sum(g) for _, g in groupby([1, 1, 2, 2, 2, 3])], len(list(product("abc", repeat=3))), sum(1 for _ in itertools.permutations("abcd")))

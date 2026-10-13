import itertools
from itertools import count, cycle, repeat, islice, chain, accumulate, takewhile, dropwhile, starmap, compress, filterfalse
print(list(islice(count(), 5)), list(islice(count(10, -3), 4)), list(islice(count(0.5, 0.25), 3)), list(islice(cycle("abc"), 7)), list(repeat("x", 3)), list(islice(repeat(1), 2)))
print(list(islice(range(20), 5)), list(islice(range(20), 2, 8)), list(islice(range(20), 2, 15, 4)), list(islice("abcdef", 3, None)), list(islice([], 3)), list(islice(count(), 0)))
it = iter(range(10))
print(list(islice(it, 3)), list(islice(it, 3)), next(it))
print(list(chain()), list(chain([1], [], [2, 3])), list(chain("ab", (1,), range(2))), list(chain.from_iterable([[1, 2], [3]])), list(itertools.chain.from_iterable(["ab", "c"])))
print(list(accumulate([1, 2, 3, 4, 5])), list(accumulate([1, 2, 3], lambda a, b: a * b)), list(accumulate("abc")), list(accumulate([], initial=5)), list(accumulate([1, 2], initial=100)), list(accumulate([3, 1, 2], max)))
print(list(takewhile(lambda x: x < 4, count())), list(dropwhile(lambda x: x < 4, range(8))), list(starmap(pow, [(2, 3), (3, 2), (10, 0)])), list(compress("abcdef", [1, 0, 1, 0, 1])), list(filterfalse(lambda x: x % 2, range(8))), list(filterfalse(None, [0, 1, "", "a"])))
for bad in (lambda: list(islice(range(5), -1)), lambda: list(islice(range(5), 0, 3, 0)), lambda: list(islice(range(5), 1, 2, 3, 4))):
    try:
        print(bad())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)

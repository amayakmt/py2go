import itertools
from itertools import groupby, zip_longest, pairwise, batched, tee
data = [("a", 1), ("a", 2), ("b", 3), ("a", 4), ("c", 5), ("c", 6)]
print([(k, [v for _, v in g]) for k, g in groupby(data, key=lambda t: t[0])])
print([(k, len(list(g))) for k, g in groupby("aaabccdd")], [k for k, _ in groupby([1, 1, 2, 3, 3, 1])], [(k, list(g)) for k, g in groupby(sorted("mississippi"))], list(groupby([])))
print(list(zip_longest("ab", [1, 2, 3])), list(zip_longest("abc", "x", fillvalue=0)), list(zip_longest()), list(zip_longest([1], [2], [3, 4], fillvalue="-")), list(zip_longest([], [])))
print(list(pairwise("abcd")), list(pairwise([1])), list(pairwise([])), list(batched("abcdefg", 3)), list(batched(range(4), 2)), list(batched([], 3)))
a, b = tee(range(3))
print(list(a), list(b), [list(x) for x in tee("ab", 3)])
try:
    list(batched("abc", 0))
except ValueError as e:
    print(e)

a: set[int] = {1, 2, 3, 4, 5}
b: set[int] = {4, 5, 6, 7}
print(a | b, a & b, a - b, b - a, a ^ b)
print(a.union(b), a.intersection(b), a.difference(b), a.symmetric_difference(b))
print(a.union([10, 11], {12}), a.intersection([1, 2, 99], [2, 3]), a.difference([1], [2]))
print(a, b)
big: set[int] = {i * 3 for i in range(20)}
small: set[int] = {0, 3, 100}
print(big & small, small & big, big - small, small - big, big | small, big ^ small)

a: set[int] = {1, 2, 3, 4}
a |= {5, 6}
print(a)
a &= {2, 3, 4, 5, 9}
print(a)
a -= {3}
print(a)
a ^= {2, 100}
print(a)
a.update([7, 8], {9})
print(a)
a.intersection_update({7, 8, 100, 5})
print(a)
a.difference_update([7])
print(a)
a.symmetric_difference_update({8, 1})
print(a)
b: set[int] = a
b.add(555)
print(a, a == b)
a.clear()
print(a, b)

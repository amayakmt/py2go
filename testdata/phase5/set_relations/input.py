a: set[int] = {1, 2}
b: set[int] = {1, 2, 3}
c: set[int] = {7}
print(a <= b, a < b, b >= a, b > a, a <= a, a < a, a >= a)
print(a == {2, 1}, a != b, b == {1, 2, 3})
print(a.issubset(b), b.issubset(a), b.issuperset(a), a.isdisjoint(c), a.isdisjoint(b))
print(a.issubset([1, 2, 5]), a.issuperset([1]), a.isdisjoint([8, 9]))
e: set[int] = set()
print(e <= a, a >= e, e < e)

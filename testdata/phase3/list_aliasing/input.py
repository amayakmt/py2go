a: list[int] = [1, 2, 3]
b: list[int] = a
b.append(4)
print(a, b, a == b)

c: list[int] = a[:]
c.append(5)
print(a, c)

d: list[int] = a.copy()
d[0] = 100
print(a, d)

e: list[int] = a + []
e.append(6)
print(a, e)

a += [7]
print(a, b)

a = a + [8]
print(a, b)

b *= 2
print(a, b)

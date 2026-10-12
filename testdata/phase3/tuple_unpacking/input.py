a: int = 1
b: int = 2
a, b = b, a
print(a, b)
x, y = 10, "ten"
print(x, y)
t: tuple[int, float] = (7, 0.5)
p, q = t
print(p, q)
(m, n), k = (1, 2), 3
print(m, n, k)
first, second, third = [1, 2, 3]
print(first, second, third)
_, last = (1, 99)
print(last)
a, b = a + b, a - b
print(a, b)
c: int = 0
d: int = 0
c, d = 5, 6
print(c, d)
u, v = divmod(17, 5)
print(u, v)

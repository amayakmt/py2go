xs: list[int] = [1, 2, 3, 4, 5]
a, *rest = xs
print(a, rest)
*initial, last = xs
print(initial, last)
first, *middle, end = xs
print(first, middle, end)
p, q, *none = [7, 8]
print(p, q, none)
t: tuple[int, int, int, int] = (1, 2, 3, 4)
h, *tail = t
print(h, tail)
c1, *cs = "hello"
print(c1, cs)
k, *vs = {10: "a", 20: "b", 30: "c"}
print(k, vs)
for head, *others in [[1, 2, 3], [4], [5, 6]]:
    print(head, others)
try:
    a1, b1, *c1x = [1]
except ValueError as e:
    print("ValueError:", e)
x1, *y1 = 1, 2, 3
print(x1, y1)

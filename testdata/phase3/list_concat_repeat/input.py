a: list[int] = [1, 2]
b: list[int] = [3]
c: list[int] = a + b
c.append(4)
print(a, b, c)
print(a * 3, 2 * b, a * 0, a * -1)
print([1, 2] + [3], [0] * 5)
e: list[int] = []
print(a + e, e + a, e + e == [])
n: int = 2
print([7, 8] * n)

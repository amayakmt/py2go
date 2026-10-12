xs: list[int] = [1, 2, 3]
for i in range(len(xs)):
    xs.append(xs[i] * 10)
print(xs)
n: int = len(xs)
for i in range(n):
    xs.pop()
print(xs)
ys: list[int] = [0, 0]
i: int = 0
while i < len(ys):
    if len(ys) < 6:
        ys.append(i)
    i += 1
print(ys, i)

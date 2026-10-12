xs: list[int] = [5, 2, 8, 1]
xs[0], xs[3] = xs[3], xs[0]
print(xs)
i: int = 1
j: int = 2
xs[i], xs[j] = xs[j], xs[i]
print(xs)
a: int = 0
xs[a], a = 100, 3
print(xs, a)
n: int = len(xs)
for p in range(n):
    for q in range(n - 1 - p):
        if xs[q] > xs[q + 1]:
            xs[q], xs[q + 1] = xs[q + 1], xs[q]
print(xs)

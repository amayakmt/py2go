xs: list[int] = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
print(xs[2:5], xs[:3], xs[-3:], xs[::-1], xs[::2], xs[1::3])
print(xs[10:], xs[-100:2], xs[5:2], xs[:100], xs[-1:-4:-1], xs[8:2:-2])
print(xs[:], xs[0:0])
a: int = 3
print(xs[a:], xs[:a], xs[a:a + 2])
ys: list[int] = xs[1:4]
ys[0] = 100
print(xs[1], ys)

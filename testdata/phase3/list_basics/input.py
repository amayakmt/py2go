xs: list[int] = [10, 20, 30]
print(xs, len(xs))
print(xs[0], xs[2], xs[-1], xs[-3])
xs[1] = 99
xs[-1] = 7
print(xs)
fs: list[float] = [1.5, 2.0, 0.1 + 0.2]
print(fs)
ss: list[str] = ["a", "b'c", 'd"e']
print(ss)
bs: list[bool] = [True, False]
print(bs)
i: int = 1
print(xs[i], xs[i + 1], xs[-i])

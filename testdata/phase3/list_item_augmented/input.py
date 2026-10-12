xs: list[int] = [1, 2, 3]
xs[0] += 10
xs[-1] *= 5
xs[1] -= 1
xs[1] **= 3
print(xs)
calls: list[int] = [0]


def pick(counter: list[int]) -> int:
    counter[0] += 1
    return 1


xs[pick(calls)] += 100
print(xs, calls)
fs: list[float] = [1.5]
fs[0] += 1
fs[0] /= 2
print(fs)
ss: list[str] = ["a"]
ss[0] += "b"
ss[0] *= 2
print(ss)

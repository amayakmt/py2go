def get(xs: list[int], i: int) -> int:
    try:
        return xs[i]
    except IndexError:
        return -1


xs: list[int] = [10, 20, 30]
print(get(xs, 1), get(xs, 9), get(xs, -3), get(xs, -4))
ys: list[int] = []
try:
    ys.pop()
except IndexError as e:
    print(e)
try:
    xs.remove(99)
except ValueError as e:
    print(e)
try:
    xs[10] = 1
except IndexError as e:
    print(e)
try:
    print(xs.index(99))
except ValueError as e:
    print(e)

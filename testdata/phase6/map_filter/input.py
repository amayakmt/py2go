xs: list[int] = [1, 2, 3, 4, 5]
print(list(map(lambda x: x * x, xs)))
print(list(filter(lambda x: x % 2 == 1, xs)))
print(list(map(lambda a, b: a + b, xs, [10, 20, 30])))
print(list(map(str, xs)), list(map(len, ["a", "bb"])), list(map(abs, [-1, 2, -3])))
print(sum(map(lambda x: x * 2, xs)), max(map(lambda x: -x, xs)))
print(list(map(lambda s: s.upper(), "abc")))
print(sorted(map(lambda x: x % 3, xs)), set(map(lambda x: x % 3, xs)))
print(dict(zip(xs, map(lambda x: x * 10, xs))))
print(len(list(filter(lambda s: len(s) > 1, ["a", "bb", "ccc"]))))
print("-".join(map(str, xs)))


def double(x: int) -> int:
    return x * 2


def is_big(x: int) -> bool:
    return x > 2


print(list(map(double, xs)), list(filter(is_big, xs)), list(map(double, filter(is_big, xs))))
print([double(x) for x in filter(is_big, xs)])

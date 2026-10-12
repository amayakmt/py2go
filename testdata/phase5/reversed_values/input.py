xs: list[int] = [1, 2, 3]
print(list(reversed(xs)), list(reversed("héy")), list(reversed(range(4))))
d: dict[str, int] = {"a": 1, "b": 2, "c": 3}
print(list(reversed(d)))
for x in reversed(xs):
    print(x)
print(sum(reversed(xs)), sorted(reversed(xs)), "".join(reversed("abc")))
print([x * 2 for x in reversed(xs)])

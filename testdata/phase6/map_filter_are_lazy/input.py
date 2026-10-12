def noisy(x: int) -> int:
    print("map", x)
    return x * 10


def keep(x: int) -> bool:
    print("filter", x)
    return x > 1


for v in map(noisy, [1, 2, 3]):
    print("got", v)
for v in filter(keep, [1, 2, 3]):
    print("kept", v)
print(any(map(lambda x: x > 1, [1, 2, 3])))
print(sum(map(noisy, [1, 2])))

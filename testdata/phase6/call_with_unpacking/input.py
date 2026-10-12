def add3(a: int, b: int, c: int) -> int:
    return a * 100 + b * 10 + c


nums: list[int] = [1, 2, 3]
print(add3(*nums))
print(add3(9, *[8, 7]))
print(add3(*(4, 5, 6)))
print(add3(*range(3)))
print(add3(*[7, 8, 9]))
opts: dict[str, int] = {"a": 1, "b": 2, "c": 3}
print(add3(**opts))
print(add3(*[1, 2], **{"c": 3}))
print(add3(1, **{"b": 5, "c": 6}))


def pack(*args: int, **kw: int) -> str:
    return str(args) + str(kw)


print(pack(*nums, **opts))
no_args: list[int] = []
no_kw: dict[str, int] = {}
print(pack(*no_args, **no_kw))
more: list[int] = [7, 8]
print(pack(0, *more))

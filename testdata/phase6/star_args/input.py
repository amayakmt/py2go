def total(*nums: int) -> int:
    t: int = 0
    for n in nums:
        t += n
    return t


def first_and_rest(first: int, *rest: int) -> str:
    return str(first) + ":" + str(len(rest)) + ":" + str(rest)


print(total(), total(1), total(1, 2, 3, 4))
print(first_and_rest(1), first_and_rest(1, 2, 3))


def describe(*items: str) -> str:
    return ",".join(items) + "|" + str(len(items)) + "|" + str(items)


print(describe(), describe("a"), describe("a", "b"))


def inspect(*xs: int) -> None:
    print(xs, len(xs), xs == (1, 2), xs[0])
    print(sorted(xs, reverse=True), sum(xs), max(xs), list(xs), 2 in xs)


inspect(1, 2)
inspect(5, 3, 9)

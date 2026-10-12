def describe(xs: list[int]) -> str:
    match xs:
        case []:
            return "empty"
        case [x]:
            return "one " + str(x)
        case [a, b]:
            return "two " + str(a + b)
        case [first, *rest] if first == 0:
            return "starts with zero, then " + str(len(rest))
        case [first, *middle, last]:
            return f"{first}..{last} with {len(middle)} between {middle}"


samples: list[list[int]] = [[], [5], [1, 2], [0, 9, 9], [1, 2, 3, 4]]
for case_ in samples:
    print(describe(case_))

pair: tuple[int, str] = (3, "x")
match pair:
    case (0, _):
        print("zero first")
    case (n, "x"):
        print("x with", n)
    case (n, s):
        print("other", n, s)
lst: list[int] = [1, 2]
match lst:
    case [1, y]:
        print("y =", y)

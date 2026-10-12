def first_negative(xs: list[int]) -> int:
    for x in xs:
        try:
            if x < 0:
                return x
        finally:
            print("checked", x)
    return 0


def safe_div(a: int, b: int) -> float:
    try:
        return a / b
    except ZeroDivisionError:
        return -1.0


def find(xs: list[str], key: str) -> int:
    i: int = 0
    while i < len(xs):
        try:
            if xs[i] == key:
                return i
        except IndexError:
            return -2
        i += 1
    return -1


def noop(flag: bool) -> None:
    try:
        if flag:
            return
        print("flag was false")
    finally:
        print("noop done")


print(first_negative([3, -4, 5]), first_negative([1]))
print(safe_div(1, 4), safe_div(1, 0))
print(find(["a", "b"], "b"), find(["a"], "z"))
noop(True)
noop(False)

def outer(n: int) -> int:
    def helper(x: int) -> int:
        return x * n

    def other(x: int) -> int:
        return helper(x) + 1

    return other(10)


print(outer(3), outer(5))


def fact(n: int) -> int:
    def walk(k: int, acc: int) -> int:
        if k <= 1:
            return acc
        return walk(k - 1, acc * k)
    return walk(n, 1)


print(fact(10))


def parity(n: int) -> bool:
    def is_even(k: int) -> bool:
        if k == 0:
            return True
        return is_odd(k - 1)

    def is_odd(k: int) -> bool:
        if k == 0:
            return False
        return is_even(k - 1)

    return is_even(n)


print(parity(10), parity(7))

def ping(n: int) -> int:
    return pong(n + 1)


def pong(n: int) -> int:
    return ping(n + 1)


def total(n: int) -> int:
    if n == 0:
        return 0
    return n + total(n - 1)


print(total(500))
try:
    ping(0)
except RecursionError as e:
    print(e)
print(total(10))

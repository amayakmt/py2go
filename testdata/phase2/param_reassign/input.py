def countdown(n: int) -> int:
    total: int = 0
    while n > 0:
        total += n
        n -= 1
    return total


def clamp(x: int, lo: int, hi: int) -> int:
    if x < lo:
        x = lo
    elif x > hi:
        x = hi
    return x


print(countdown(4))
print(clamp(15, 0, 10))
print(clamp(-5, 0, 10))
print(clamp(5, 0, 10))

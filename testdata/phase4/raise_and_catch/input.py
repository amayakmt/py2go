def check(n: int) -> int:
    if n < 0:
        raise ValueError("negative")
    return n * 2


try:
    print(check(2))
    print(check(-1))
    print("not reached")
except ValueError as e:
    print("caught", e)
print("after")

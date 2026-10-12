def find_first_multiple(k: int, limit: int) -> int:
    for i in range(1, limit):
        if i % k == 0 and i > 10:
            return i
    return -1


def spin(n: int) -> int:
    i: int = 0
    while True:
        i += 1
        if i * i > n:
            return i


print(find_first_multiple(7, 100))
print(find_first_multiple(7, 12))
print(spin(50))

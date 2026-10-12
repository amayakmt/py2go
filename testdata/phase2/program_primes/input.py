def is_prime(n: int) -> bool:
    if n < 2:
        return False
    i: int = 2
    while i * i <= n:
        if n % i == 0:
            return False
        i += 1
    return True


count: int = 0
for n in range(100):
    if is_prime(n):
        count += 1
        print(n)
print("count", count)

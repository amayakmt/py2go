def sieve(n: int) -> list[int]:
    is_prime: list[bool] = [True] * (n + 1)
    is_prime[0] = False
    is_prime[1] = False
    for i in range(2, n + 1):
        if is_prime[i]:
            for j in range(i * i, n + 1, i):
                is_prime[j] = False
    primes: list[int] = []
    for i in range(n + 1):
        if is_prime[i]:
            primes.append(i)
    return primes


ps: list[int] = sieve(60)
print(ps)
print(len(ps), sum(ps))

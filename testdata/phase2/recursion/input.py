def fact(n: int) -> int:
    if n <= 1:
        return 1
    return n * fact(n - 1)


def fib(n: int) -> int:
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)


def gcd(a: int, b: int) -> int:
    while b != 0:
        t: int = b
        b = a % b
        a = t
    return a


print(fact(10))
print(fact(20))
print(fib(15))
print(gcd(48, 18))

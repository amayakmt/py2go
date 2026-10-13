def gcd(a, b):
    while b:
        a, b = b, a % b
    return abs(a)

def egcd(a, b):
    if b == 0:
        return (a, 1, 0)
    g, x, y = egcd(b, a % b)
    return (g, y, x - (a // b) * y)

def is_probable_prime(n):
    if n < 2:
        return False
    for p in (2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37):
        if n % p == 0:
            return n == p
    d, s = n - 1, 0
    while d % 2 == 0:
        d //= 2
        s += 1
    for a in (2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37):
        x = pow(a, d, n)
        if x in (1, n - 1):
            continue
        for _ in range(s - 1):
            x = x * x % n
            if x == n - 1:
                break
        else:
            return False
    return True

print(gcd(2 ** 100, 6 ** 50), gcd(-12, 18), gcd(0, 0), gcd(10 ** 30 + 1, 10 ** 15 + 1), egcd(240, 46), egcd(2 ** 70, 3 ** 40))
print([n for n in range(2, 60) if is_probable_prime(n)])
print([is_probable_prime(n) for n in (2 ** 61 - 1, 2 ** 89 - 1, 2 ** 127 - 1, 2 ** 67 - 1, 10 ** 18 + 9, 10 ** 18 + 7, 1000000007 * 998244353)])
n = 10 ** 30
while not is_probable_prime(n):
    n += 1
print(n, n - 10 ** 30)

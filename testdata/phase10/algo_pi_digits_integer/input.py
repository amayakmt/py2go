def arctan_inv(x, digits):
    unity = 10 ** (digits + 10)
    total = term = unity // x
    x2 = x * x
    n, sign = 3, -1
    while term:
        term //= x2
        total += sign * (term // n)
        sign = -sign
        n += 2
    return total

def pi_machin(digits):
    unity = 10 ** (digits + 10)
    pi = 4 * (4 * arctan_inv(5, digits) - arctan_inv(239, digits))
    return pi // 10 ** 10

p = pi_machin(100)
print(p)
print(len(str(pi_machin(500))), str(pi_machin(500))[-10:])
def e_digits(n):
    unity = 10 ** (n + 5)
    total, term, k = unity, unity, 1
    while term:
        term //= k
        total += term
        k += 1
    return total // 10 ** 5
print(e_digits(60))
def isqrt(n):
    x = n
    y = (x + 1) // 2
    while y < x:
        x, y = y, (y + n // y) // 2
    return x
print(isqrt(2 * 10 ** 100), isqrt(10 ** 200) == 10 ** 100)

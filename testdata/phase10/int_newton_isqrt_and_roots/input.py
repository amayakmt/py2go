def isqrt(n):
    if n < 0:
        raise ValueError("square root of negative number")
    if n == 0:
        return 0
    x = 1 << ((n.bit_length() + 1) // 2)
    while True:
        y = (x + n // x) // 2
        if y >= x:
            return x
        x = y

for n in (0, 1, 2, 3, 4, 15, 16, 17, 10 ** 10, 2 ** 64, 2 ** 64 - 1, 10 ** 40, 10 ** 41, 2 ** 200 + 1):
    r = isqrt(n)
    print(n, r, r * r <= n < (r + 1) * (r + 1))
def iroot(n, k):
    lo, hi = 0, 1 << (n.bit_length() // k + 1)
    while lo < hi:
        mid = (lo + hi + 1) // 2
        if mid ** k <= n:
            lo = mid
        else:
            hi = mid - 1
    return lo
print(iroot(10 ** 30, 3), iroot(2 ** 100, 5), iroot(3 ** 90, 9), iroot(10 ** 100 + 1, 10))
try:
    isqrt(-1)
except ValueError as e:
    print(e)
# sqrt(2) to 60 digits with integers
print(isqrt(2 * 10 ** 120))

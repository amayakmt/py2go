import math
print(math.factorial(0), math.factorial(5), math.factorial(20), math.factorial(25), len(str(math.factorial(1000))))
print(math.comb(5, 2), math.comb(50, 25), math.comb(10, 0), math.comb(3, 5), math.perm(5, 2), math.perm(5), math.perm(10, 10), math.perm(3, 5))
print(math.gcd(12, 18), math.gcd(0, 5), math.gcd(0, 0), math.gcd(-12, 18), math.gcd(2 ** 70, 6 ** 30), math.gcd(), math.gcd(12, 18, 24), math.lcm(4, 6), math.lcm(3, 0), math.lcm(), math.lcm(2 ** 40, 3 ** 20, 5))
print(math.isqrt(0), math.isqrt(15), math.isqrt(16), math.isqrt(10 ** 40), math.isqrt(2 ** 200 + 1), math.prod([1, 2, 3, 4]), math.prod([]), math.prod([0.5, 4]), math.prod(range(1, 21)), math.prod([2, 3], start=10))
print(math.fsum([0.1] * 10), math.fsum([1e100, 1.0, -1e100]), math.fsum([]), math.fsum([1, 2, 3]), math.sumprod([1, 2, 3], [4, 5, 6]), math.sumprod([0.5], [4]))
for f, args in ((math.factorial, (1.5,)), (math.factorial, ("a",)), (math.comb, (-1, 2)), (math.comb, (3, -1)), (math.perm, (-1,)), (math.gcd, (1.5,)), (math.isqrt, (2.5,)), (math.lcm, ("a",)), (math.sumprod, ([1], [1, 2]))):
    try:
        print(f(*args))
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)

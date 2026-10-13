def fib(n):
    a, b = 0, 1
    for _ in range(n):
        a, b = b, a + b
    return a

def fact(n):
    r = 1
    for i in range(2, n + 1):
        r *= i
    return r

def fib_fast(n):
    def go(k):
        if k == 0:
            return (0, 1)
        a, b = go(k >> 1)
        c = a * (2 * b - a)
        d = a * a + b * b
        return (d, c + d) if k & 1 else (c, d)
    return go(n)[0]

print(fib(10), fib(90), fib(100), fib(300), fib_fast(1000) == fib(1000), len(str(fib(5000))), fib_fast(10 ** 5) % 10 ** 12)
print(fact(20), fact(21), fact(30), len(str(fact(1000))), fact(100) // fact(98), sum(map(int, str(fact(100)))), fact(25) % 10 ** 9 + 7)
print(fib(200) // fib(100), fib(201) * fib(199) - fib(200) ** 2, [fib(i) for i in range(60, 66)], str(fact(60))[:20], str(fact(60)).count("0"))
tz = 0
n = fact(1000)
while n % 10 == 0:
    n //= 10
    tz += 1
print(tz, n % 1000)

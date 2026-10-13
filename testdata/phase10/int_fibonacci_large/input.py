a, b = 0, 1
for i in range(1, 301):
    a, b = b, a + b
    if i in (50, 90, 93, 94, 100, 150, 300):
        print(i, a)
print(len(str(a)), a % 97, a // 10 ** 50, str(a)[-5:])
def fib(n, memo={}):
    if n < 2:
        return n
    if n not in memo:
        memo[n] = fib(n - 1) + fib(n - 2)
    return memo[n]
print(fib(200), fib(10), fib(92) + fib(93))

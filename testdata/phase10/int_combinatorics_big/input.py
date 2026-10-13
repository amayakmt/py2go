def binom(n, k):
    r = 1
    for i in range(1, k + 1):
        r = r * (n - k + i) // i
    return r

def catalan(n):
    return binom(2 * n, n) // (n + 1)

def stirling2(n, k):
    if n == k:
        return 1
    if k == 0 or k > n:
        return 0
    return k * stirling2(n - 1, k) + stirling2(n - 1, k - 1)

print([binom(10, k) for k in range(11)], binom(100, 50), binom(1000, 500) % 10 ** 9, len(str(binom(1000, 500))))
print([catalan(n) for n in range(15)], catalan(100))
row = [1]
for _ in range(60):
    row = [1] + [a + b for a, b in zip(row, row[1:])] + [1]
print(row[28:33], sum(row) == 2 ** 60)
print([stirling2(8, k) for k in range(9)], sum(stirling2(10, k) for k in range(11)))
def partitions(n, memo={}):
    if n == 0:
        return 1
    if n in memo:
        return memo[n]
    total, k = 0, 1
    while k * (3 * k - 1) // 2 <= n:
        sign = 1 if k % 2 else -1
        total += sign * partitions(n - k * (3 * k - 1) // 2)
        if k * (3 * k + 1) // 2 <= n:
            total += sign * partitions(n - k * (3 * k + 1) // 2)
        k += 1
    memo[n] = total
    return total
print([partitions(n) for n in range(15)], partitions(200), partitions(500))

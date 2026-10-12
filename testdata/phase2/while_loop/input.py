i: int = 0
total: int = 0
while i < 5:
    total += i
    i += 1
print(total)
print(i)
n: int = 27
steps: int = 0
while n != 1:
    if n % 2 == 0:
        n = n // 2
    else:
        n = 3 * n + 1
    steps += 1
print(steps)

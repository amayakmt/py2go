def find_prime_factor(n: int) -> int:
    for d in range(2, n):
        if n % d == 0:
            return d
    else:
        return n
    return 0


print([find_prime_factor(k) for k in [7, 12, 15, 13]])
for i in range(3):
    if i == 5:
        break
else:
    print("no break in range")
for i in range(3):
    if i == 1:
        break
else:
    print("never printed")
n: int = 0
while n < 3:
    n += 1
else:
    print("while finished", n)
while True:
    n += 1
    if n > 5:
        break
else:
    print("never")
print(n)
rows: list[list[int]] = [[1, 2], [3, 4]]
for row in rows:
    for v in row:
        if v == 3:
            break
    else:
        continue
    print("found 3 in", row)
    break
else:
    print("not found")
empty: list[int] = []
for k in empty:
    pass
else:
    print("empty loop runs else")

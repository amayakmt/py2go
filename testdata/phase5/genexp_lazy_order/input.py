def noisy(n: int) -> int:
    print("produce", n)
    return n * n


for v in (noisy(i) for i in range(3)):
    print("consume", v)
total: int = 0
for v in (noisy(i) for i in range(2) if i != 5):
    total += v
print(total)

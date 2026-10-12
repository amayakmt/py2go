x: int = 100
print([x for x in range(3)], x)
y: str = "outer"
print([y * 2 for y in "ab"], y)
n: int = 5
print([i + n for i in range(3)])
total: int = 0
print([total + i for i in range(3)], total)
k: int = 7
print({k: k for k in range(2)}, k)
print({k for k in range(2)}, k)
print(sum(k for k in range(4)), k)


def f(limit: int) -> list[int]:
    return [i for i in range(limit) if i < limit - 1]


print(f(4))

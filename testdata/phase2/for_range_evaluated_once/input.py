n: int = 3
for i in range(n):
    n += 10
    print(i, n)
print(n)
calls: int = 0


def bound() -> int:
    print("bound called")
    return 2


for i in range(bound()):
    print(i)

results: list[int] = []
for d in [2, 0, 5, 0, 1]:
    try:
        results.append(100 // d)
    except ZeroDivisionError:
        results.append(-1)
print(results)

count: int = 0
n: int = 10
while n > -3:
    try:
        count += 12 // n
    except ZeroDivisionError:
        print("hit zero")
    n -= 1
print(count)

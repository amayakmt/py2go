def run(n: int) -> int:
    total: int = 0
    for i in range(n):
        try:
            total += 10 // (2 - i)
        finally:
            print("step", i)
    return total


try:
    print(run(2))
    print(run(4))
except ZeroDivisionError as e:
    print("stopped:", e)

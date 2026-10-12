def process(items: list[int]) -> int:
    total: int = 0
    for x in items:
        if x == 0:
            raise ZeroDivisionError("zero item")
        total += 100 // x
    return total


print(process([1, 2, 5]))
try:
    print(process([4, 0, 5]))
except ZeroDivisionError as e:
    print("loop aborted:", e)

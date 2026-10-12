items: list[str] = ["1", "x", "30", "", "4.5", "-7"]
total: int = 0
bad: int = 0
for s in items:
    try:
        total += int(s)
    except ValueError as e:
        bad += 1
        print("skip", repr(s), "-", e)
print(total, bad)

matrix: list[list[int]] = [[1, 2], [3], [], [4, 5]]
for row in matrix:
    try:
        print(row[0] + row[1])
    except IndexError:
        print("short row", row)

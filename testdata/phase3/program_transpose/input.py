def transpose(m: list[list[int]]) -> list[list[int]]:
    rows: int = len(m)
    cols: int = len(m[0])
    out: list[list[int]] = []
    for c in range(cols):
        row: list[int] = []
        for r in range(rows):
            row.append(m[r][c])
        out.append(row)
    return out


m: list[list[int]] = [[1, 2, 3], [4, 5, 6]]
t: list[list[int]] = transpose(m)
print(t)
for row in t:
    print(",".join([str(row[0]), str(row[1])]))

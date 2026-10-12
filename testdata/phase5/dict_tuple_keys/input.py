grid: dict[tuple[int, int], str] = {}
grid[(0, 0)] = "origin"
grid[(1, 2)] = "p"
grid[(1, 2)] += "!"
print(grid, grid[(1, 2)], (0, 0) in grid, (5, 5) in grid)
for (x, y), v in grid.items():
    print(x, y, v)
memo: dict[tuple[int, int], int] = {}


def paths(r: int, c: int, memo: dict[tuple[int, int], int]) -> int:
    if r == 0 or c == 0:
        return 1
    if (r, c) in memo:
        return memo[(r, c)]
    memo[(r, c)] = paths(r - 1, c, memo) + paths(r, c - 1, memo)
    return memo[(r, c)]


print(paths(10, 10, memo), len(memo))

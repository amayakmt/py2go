empty: list[list[int]] = [[] for _ in range(3)]
empty[0].append(1)
print(empty)
grid: list[list[int]] = [[0] * 3 for _ in range(2)]
grid[0][0] = 5
print(grid)
maps: list[dict[str, int]] = [{} for _ in range(2)]
maps[1]["k"] = 1
print(maps)
none: list[int] = [x for x in range(0)]
print(none, len(none))

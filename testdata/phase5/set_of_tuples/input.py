visited: set[tuple[int, int]] = set()
for x in range(3):
    for y in range(3):
        if (x + y) % 2 == 0:
            visited.add((x, y))
print(sorted(visited), len(visited))
print((1, 1) in visited, (0, 1) in visited)
pairs: set[tuple[str, int]] = {("a", 1), ("b", 2)}
print(sorted(pairs))

grid = [[0] * 3 for _ in range(3)]
grid[1][1] = 5
grid[0].append(7)
print(grid)
aliased = [[0] * 2] * 2
aliased[0][0] = 9
print(aliased)
data = {"users": [{"name": "a", "tags": []}, {"name": "b", "tags": ["x"]}]}
data["users"][0]["tags"].append("new")
data["users"].append({"name": "c", "tags": None})
for u in data["users"]:
    print(u["name"], u["tags"])
copy = [row[:] for row in grid]
copy[1][1] = -1
print(grid[1][1], copy[1][1])
x = y = []
x.append(1)
print(y, x is y)
a = [1, 2, 3]
b = a
b += [4]
c = a + [5]
print(a, b, c, a is b, a is c)

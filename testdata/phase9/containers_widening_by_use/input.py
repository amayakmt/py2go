xs = [1, 2, 3]
xs.append("four")
xs[0] = 2.5
xs[1:2] = [None, True]
print(xs, len(xs), xs.count(None))

d = {"a": 1}
d["b"] = "two"
d[3] = [3]
d[(4, 5)] = 4.5
print(d, sorted(map(str, d)))

s = {1}
s.add("x")
s.add(2.5)
s.add((1, 2))
print(len(s), "x" in s, (1, 2) in s, 2.5 in s)

def collect(n):
    out = []
    for i in range(n):
        out.append(i if i % 2 else str(i))
    return out

def make_grid(rows, cols):
    grid = [[0] * cols for _ in range(rows)]
    grid[0][0] = 0.5
    grid[rows - 1][cols - 1] = "end"
    return grid

print(collect(5), make_grid(2, 3))
nested = [[1, 2], [3]]
alias = nested
alias[0].append(4)
print(nested, nested is alias, nested[0] is alias[0])

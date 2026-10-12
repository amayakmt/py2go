data = [1, "a", 2.5, None, 3, "b", [1], (2,)]
print([x for x in data if isinstance(x, int)])
print([str(x) for x in data])
print({type(x).__name__ for x in data} == {"int", "str", "float", "NoneType", "list", "tuple"})
print({x: type(x).__name__ for x in data if isinstance(x, (int, str))})
print([(i, j) for i in range(3) for j in range(i)])
print([[c * n for c in "ab"] for n in (1, 2)])
squares = (n * n for n in range(5))
print(next(squares), next(squares), list(squares))
print(sum(len(str(x)) for x in data), max(len(str(x)) for x in data))
flat = [y for x in [[1, 2], [3], []] for y in x]
print(flat, [x for x in flat if x % 2], [x if x > 1 else None for x in flat])
matrix = [[1, 2, 3], [4, 5, 6]]
print([list(r) for r in zip(*matrix)], [sum(r) for r in matrix], [r[::-1] for r in matrix])

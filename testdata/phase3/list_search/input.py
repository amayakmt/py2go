xs: list[int] = [5, 3, 5, 1, 5]
print(5 in xs, 4 in xs, 4 not in xs, 1 not in xs)
print(xs.index(3), xs.index(5), xs.count(5), xs.count(9))
names: list[str] = ["ann", "bob"]
print("bob" in names, "carl" in names)
pairs: list[tuple[int, str]] = [(1, "a"), (2, "b")]
print((2, "b") in pairs, (2, "a") in pairs)
print(2 in (1, 2, 3), 5 in (1, 2, 3), 5 not in (1, 2, 3))
print("x" in ("x", "y"))

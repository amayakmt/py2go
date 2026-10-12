print([[1, 2], [3]], ([1], (2, "x")), [(1, 2.0)], ["it's", 'say "hi"', "both ' \""])
print(["tab\t", "nl\n", "é", "\x00"])
e1: list[int] = []
e2: list[list[int]] = [[]]
print(e1, e2, (1,))
print([1.0, 1e16, 1e-5, -0.0], [float("inf"), float("-inf")])
print((1, [(2, 3)], ([4],)))
x: list[str] = ["a"]
print(f"{x} {x!r} {x!s}")
print(str(x) + "!", repr(x), len(repr(x)))

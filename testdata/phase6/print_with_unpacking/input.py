items: list[int] = [1, 2, 3]
print(*items)
print(*items, sep="-")
print("start", *items, "end")
print(*items, *items, sep=",", end="!\n")
words: list[str] = ["a", "b"]
print(*words)
print(*"xyz")
nothing: list[int] = []
print(*nothing, sep="|")
print("only")

d: dict[int, str] = {1: "a", 2: "b", 3: "c", 4: "d"}
del d[2]
print(d)
d[2] = "B"
print(d)
del d[1]
d[1] = "A"
print(d)
d[3] = "C"
print(d)
print(list(d), list(d.values()))
d.clear()
print(d)
d[9] = "z"
print(d)

d = {}
d[1] = "int"
d[1.0] = "float"
d[True] = "bool"
print(d, len(d))
s = {1, 1.0, True, 2, 2.0}
print(s, len(s))
print(1 in {1.0}, 1.5 in {1.5}, 2 in {2.0: 0})
d2 = {1: "a", 2: "b"}
d2[1.0] = "A"
print(d2, [type(k).__name__ for k in d2])

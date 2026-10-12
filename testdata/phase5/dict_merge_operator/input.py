a: dict[int, str] = {1: "a", 2: "b"}
b: dict[int, str] = {2: "B", 3: "C"}
c: dict[int, str] = a | b
print(c, a, b)
a |= b
print(a)
d: dict[int, str] = {0: "zero"} | {1: "one"}
print(d)

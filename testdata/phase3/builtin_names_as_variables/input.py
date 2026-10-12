max: int = 5
len: int = 2
sum: int = 7
list: str = "not a type"
for c in "ab":
    print(c, max, len, sum, list)
for i, c in enumerate("xy"):
    print(i, c)
for x in [1, 2]:
    print(x + max)

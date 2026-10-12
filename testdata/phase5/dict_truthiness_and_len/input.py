d: dict[int, int] = {}
if not d:
    print("empty")
d[1] = 1
if d:
    print("not empty", len(d))
while d:
    d.popitem()
print(len(d))

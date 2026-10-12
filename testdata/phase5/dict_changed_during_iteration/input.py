d: dict[int, int] = {1: 1, 2: 2, 3: 3}
print("before")
for k in d:
    d[k + 10] = 0
print("never")

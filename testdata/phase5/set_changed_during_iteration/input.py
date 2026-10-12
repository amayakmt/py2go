s: set[int] = {1, 2, 3}
print("before")
for x in s:
    s.add(x + 10)
print("never")

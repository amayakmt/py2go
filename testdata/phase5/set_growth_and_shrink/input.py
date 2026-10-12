s: set[int] = set()
for i in range(100):
    s.add(i * 37 % 211)
print(list(s))
for i in range(0, 100, 3):
    s.discard(i * 37 % 211)
print(list(s), len(s))
for i in range(200, 220):
    s.add(i)
print(list(s))
while len(s) > 5:
    s.pop()
print(s)

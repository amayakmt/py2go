s: set[int] = {5, 3, 9, 1, 7, 2, 8}
out: list[int] = []
while s:
    out.append(s.pop())
print(out)
t: set[int] = {i * 11 for i in range(12)}
print(t.pop(), t.pop(), t.pop(), sorted(t))
t.add(1000)
print(t.pop(), len(t))

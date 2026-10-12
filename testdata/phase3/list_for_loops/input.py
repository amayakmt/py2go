xs: list[int] = [3, 1, 4]
for x in xs:
    print(x)
print(x)
total: int = 0
for v in xs:
    total += v
print(total)
for c in "héy":
    print(c)
for w in "a b c".split():
    print(w.upper())
for row in [[1, 2], [3]]:
    print(len(row), row)
for t in [(1, "a"), (2, "b")]:
    print(t, t[0], t[1])
empty: list[int] = []
for q in empty:
    print("never", q)

names: list[str] = ["ann", "bob", "cy"]
for i, n in enumerate(names):
    print(i, n)
for i, n in enumerate(names, 1):
    print(f"{i}. {n}")
start: int = 10
for i, c in enumerate("hé", start):
    print(i, c)
ages: list[int] = [30, 25, 41]
for n, a in zip(names, ages):
    print(n, a)
for a, b, flag in zip([1, 2, 3], "xyz", [True, False, True]):
    print(a, b, flag)
for a, b in zip([1, 2, 3, 4], ["a", "b"]):
    print(a, b)
for p in zip([1, 2], [3, 4]):
    print(p)
for e in enumerate(names):
    print(e)
for i, (m, k) in enumerate(zip([1, 2], [3, 4])):
    print(i, m, k)
for _, n in enumerate(names):
    print(n)
for x, _ in zip(names, ages):
    print(x)
print(i, a, b, m, k)

groups: dict[str, list[int]] = {}
for n in range(10):
    key: str = "x"
    if n % 2 == 0:
        key = "even"
    else:
        key = "odd"
    groups.setdefault(key, []).append(n)
print(groups)
d: dict[str, int] = {}
print(d.setdefault("a", 1), d.setdefault("a", 2), d)

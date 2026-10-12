d: dict[str, int] = {"x": 1, "y": 2, "z": 3}
for k, v in d.items():
    print(k, v)
for k in d:
    print(k, d[k])
for v in d.values():
    print(v * 10)
for k in d.keys():
    print(k.upper())
total: int = 0
for _, v in d.items():
    total += v
print(total)
for i, (k, v) in enumerate(d.items()):
    print(i, k, v)

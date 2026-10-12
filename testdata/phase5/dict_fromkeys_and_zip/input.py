print(dict.fromkeys(["a", "b", "a"], 0))
print(dict.fromkeys(range(3), "x"))
names: list[str] = ["ann", "bob", "cy"]
ages: list[int] = [30, 25, 41]
print(dict(zip(names, ages)))
print(dict(zip(range(3), names)))
print(dict([(1, "one"), (2, "two")]))
print(dict(a=1, b=2))
print({k: v for k, v in zip(names, ages)})

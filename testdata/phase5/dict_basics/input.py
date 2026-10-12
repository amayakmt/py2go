d: dict[str, int] = {"one": 1, "two": 2, "three": 3}
print(d, len(d))
print(d["two"], d.get("four", -1), d.get("one", -1))
d["four"] = 4
d["one"] = 100
print(d)
print("two" in d, "five" in d, "five" not in d)
e: dict[int, str] = {}
print(e, len(e), bool(e), bool(d))
e[3] = "c"
e[1] = "a"
print(e)
f: dict[bool, int] = {True: 1, False: 0}
print(f, f[True])

d = {}
d["a"] = 1
d[2] = "two"
d[3.5] = [1, 2]
d[(1, 2)] = {"nested": True}
d[None] = "none key"
d[True] = "true key"
print(d)
print(len(d), d["a"], d[2], d[3.5], d[(1, 2)]["nested"], d[None], d[1])
print(d.get("missing"), d.get("missing", 0), d.pop(2), d.pop("zzz", "dflt"))
d.setdefault("new", []).append(1)
d.setdefault("new", []).append(2)
print(d["new"], list(d.keys()), list(d.values())[:2])
d.update({"x": 1}, y=2)
print(sorted(map(str, d)), "x" in d, 99 in d)
for k, v in list(d.items())[:3]:
    print(repr(k), "->", repr(v))
del d["a"]
print("a" in d, len(d))
copy = dict(d)
copy["added"] = 1
print("added" in d, "added" in copy, d == dict(d))

from collections import defaultdict, Counter
dd = defaultdict(list)
for k, v in [("a", 1), ("b", 2), ("a", 3)]:
    dd[k].append(v)
print(dict(dd) if False else sorted(dd.items()), len(dd), "z" in dd, dd["z"], "z" in dd, list(dd), sorted(dd.keys()), dd.get("q"), dd.get("a"))
di = defaultdict(int)
for ch in "hello":
    di[ch] += 1
print(sorted(di.items()), di["l"], di.pop("h"), "h" in di, defaultdict(lambda: "none")["x"], defaultdict(set)["s"], repr(defaultdict(int)).startswith("defaultdict("))
nested = defaultdict(lambda: defaultdict(int))
nested["a"]["x"] += 1; nested["a"]["y"] += 2; nested["b"]["x"] += 3
print(sorted((k, sorted(v.items())) for k, v in nested.items()))
try:
    defaultdict()["missing"]
except KeyError as e:
    print("KeyError", e)
c = Counter("abracadabra")
print(c, c.most_common(2), c.most_common(), c["a"], c["zzz"], len(c), sorted(c), sorted(c.elements()), c.total(), c.most_common(0))
c.update("aab"); c.subtract("bbb"); c["new"] += 2; del c["r"]
print(c, sorted(c.items()), Counter({"x": 2, "y": -1}), Counter(a=3, b=1), Counter([1, 1, 2]), Counter())
a, b = Counter("aabbc"), Counter("abbd")
print(a + b, a - b, a & b, a | b, (a + b) == Counter("aabbcabbd"), sum(a.values()), a.copy() == a)
print(Counter("hello world").most_common(3), [k for k, _ in Counter("zzyxx").most_common()], max(Counter("aabbb").items(), key=lambda kv: kv[1]))

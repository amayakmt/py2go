import copy
class Box:
    def __init__(self, items, label):
        self.items = items
        self.label = label
    def __repr__(self):
        return f"Box({self.items}, {self.label!r})"
a = [1, [2, 3], {"k": [4]}, (5, [6]), {7, 8}, "s", None, 1.5]
c = copy.copy(a)
d = copy.deepcopy(a)
a[1].append(99); a[2]["k"].append(100); a[3][1].append(101)
print(a, c, d, c is a, c[1] is a[1], d[1] is a[1], d[3][1] is a[3][1])
b = Box([1, 2], "x")
bc = copy.copy(b)
bd = copy.deepcopy(b)
b.items.append(3); b.label = "changed"
print(b, bc, bd, bc.items is b.items, bd.items is b.items)
shared = [0]
pair = [shared, shared]
dp = copy.deepcopy(pair)
print(dp[0] is dp[1], dp[0] is shared, dp)
cyc = [1]
cyc.append(cyc)
dc = copy.deepcopy(cyc)
print(dc[1] is dc, dc[0], len(dc))
d1 = {"a": {"b": [1]}}
d2 = copy.deepcopy(d1); d1["a"]["b"].append(2)
print(d1, d2, copy.copy("immutable") == "immutable", copy.deepcopy(5), copy.deepcopy((1, 2)), copy.copy(None))
print(copy.deepcopy(frozenset([1, 2])) == frozenset([1, 2]), copy.copy(range(3)), copy.deepcopy([x * x for x in range(4)]))

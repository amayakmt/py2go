import json


class Counter(dict):
    def __missing__(self, key):
        return 0

    def add(self, k):
        self[k] = self[k] + 1


class Logged(dict):
    def __init__(self, *args, **kw):
        super().__init__(*args, **kw)
        self.log = []

    def __setitem__(self, key, value):
        self.log.append(key)
        dict.__setitem__(self, key, value)


class Stack(list):
    def push(self, x):
        self.append(x)

    def peek(self):
        return self[-1]


class Tags(set):
    pass


c = Counter()
c.add("a")
c.add("a")
c.add("b")
print(c, c["zzz"], len(c), "a" in c, sorted(c.items()), isinstance(c, dict), type(c).__name__)
lg = Logged(x=1)
lg["y"] = 2
lg.update(z=3)
print(lg, lg.log, dict(lg), json.dumps(lg))
s = Stack([1, 2])
s.push(3)
print(s, s.peek(), len(s), s[0:2], s + [9], isinstance(s, list), list(reversed(s)), sum(s))
s += [4]
s.sort(reverse=True)
print(s, s == [4, 3, 2, 1], ", ".join(str(v) for v in s))
t = Tags([1, 2])
t.add(3)
print(sorted(t), 2 in t, t | {9} == {1, 2, 3, 9}, isinstance(t, set), len(t))
def f(**kw):
    return sorted(kw)
print(f(**lg), {**c})


class OrderedSet(list):
    def __init__(self, items=()):
        super().__init__()
        for i in items:
            if i not in self:
                self.append(i)


o = OrderedSet([3, 1, 3, 2, 1])
print(o, o.count(3), o.index(2), o * 2, 2 * o, o[::-1])
try:
    hash(o)
except TypeError as e:
    print("TypeError", e)
c2 = Counter(a=5)
c2 |= {"b": 1}
print(c2, c2 == {"a": 5, "b": 1}, c2.copy(), type(c2.copy()).__name__, c2.get("q", "none"), c2.pop("a"), c2)
print(isinstance(c2, (list, dict)), isinstance(s, dict), isinstance(t, (list, tuple)))

class Stack:
    def __init__(self, *items):
        self.items = list(items)

    def __len__(self):
        return len(self.items)

    def __getitem__(self, i):
        return self.items[i]

    def __setitem__(self, i, v):
        self.items[i] = v

    def __delitem__(self, i):
        del self.items[i]

    def __contains__(self, v):
        return v in self.items

    def __iter__(self):
        return iter(reversed(self.items))

    def __bool__(self):
        return bool(self.items)

    def __repr__(self):
        return "Stack(%s)" % ", ".join(map(repr, self.items))

s = Stack(1, "two", 3.0)
print(len(s), s[0], s[-1], s[1:], 3.0 in s, "x" in s, bool(s), bool(Stack()))
s[1] = "TWO"
del s[0]
print(s, list(s), [x for x in s], sorted(map(str, s)))
for x in s:
    print("item", x)
a, b = s
print(a, b)
print(list(enumerate(s)), "yes" if s else "no", "yes" if Stack() else "no")

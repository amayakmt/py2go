class P:
    def __init__(self, v):
        self.v = v

    def __eq__(self, other):
        return self.v == other.v

    def __hash__(self):
        return hash(self.v)

d = {P(1): "a"}
d[P(1)] = "b"
d[P(2)] = "c"
print(len(d), d[P(1)], P(2) in d, sorted(d.values()), len({P(1), P(1), P(3)}))

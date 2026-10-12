class P:
    def __init__(self, v):
        self.v = v

    def __eq__(self, other):
        return self.v == other.v

    def __hash__(self):
        return hash(self.v)

d = {P(1): "a"}

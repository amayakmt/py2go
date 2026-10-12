class Node:
    def accept(self, v: "Visitor") -> int:
        return 0


class Num(Node):
    def __init__(self, n: int) -> None:
        self.n = n

    def accept(self, v: "Visitor") -> int:
        return v.visit_num(self)


class Add(Node):
    def __init__(self, l: Node, r: Node) -> None:
        self.l = l
        self.r = r

    def accept(self, v: "Visitor") -> int:
        return v.visit_add(self)


class Mul(Node):
    def __init__(self, l: Node, r: Node) -> None:
        self.l = l
        self.r = r

    def accept(self, v: "Visitor") -> int:
        return v.visit_mul(self)


class Visitor:
    def visit_num(self, n: Num) -> int:
        return n.n

    def visit_add(self, a: Add) -> int:
        return a.l.accept(self) + a.r.accept(self)

    def visit_mul(self, m: Mul) -> int:
        return m.l.accept(self) * m.r.accept(self)


class Counter(Visitor):
    def __init__(self) -> None:
        self.nodes = 0

    def visit_num(self, n: Num) -> int:
        self.nodes += 1
        return 1

    def visit_add(self, a: Add) -> int:
        self.nodes += 1
        return a.l.accept(self) + a.r.accept(self)

    def visit_mul(self, m: Mul) -> int:
        self.nodes += 1
        return m.l.accept(self) + m.r.accept(self)


tree: Node = Add(Num(2), Mul(Num(3), Add(Num(4), Num(5))))
print(tree.accept(Visitor()))
c: Counter = Counter()
print(tree.accept(c), c.nodes)

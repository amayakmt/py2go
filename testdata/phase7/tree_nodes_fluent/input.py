class Node:
    def __init__(self, v: int) -> None:
        self.v = v
        self.kids: list[Node] = []

    def add(self, k: "Node") -> "Node":
        self.kids.append(k)
        return self

    def total(self) -> int:
        return self.v + sum(k.total() for k in self.kids)

    def depth(self) -> int:
        best: int = 0
        for k in self.kids:
            best = max(best, k.depth())
        return best + 1

    def flatten(self) -> list[int]:
        out: list[int] = [self.v]
        for k in self.kids:
            out += k.flatten()
        return out


root: Node = Node(1).add(Node(2).add(Node(4))).add(Node(3))
print(root.total(), root.depth(), root.flatten())
print(root.kids[0].kids[0].v, len(root.kids))

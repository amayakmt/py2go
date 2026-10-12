class Node:
    def __init__(self, value: int, next_node: "Node") -> None:
        self.value = value
        self.next = next_node


class Tree:
    def __init__(self, value: int) -> None:
        self.value = value
        self.left: list[Tree] = []
        self.right: list[Tree] = []

    def insert(self, v: int) -> None:
        side: list[Tree] = self.right
        if v < self.value:
            if len(self.left) == 0:
                self.left.append(Tree(v))
            else:
                self.left[0].insert(v)
        else:
            if len(side) == 0:
                side.append(Tree(v))
            else:
                side[0].insert(v)

    def walk(self, out: list[int]) -> None:
        if len(self.left) > 0:
            self.left[0].walk(out)
        out.append(self.value)
        if len(self.right) > 0:
            self.right[0].walk(out)

    def depth(self) -> int:
        l: int = 0
        r: int = 0
        if len(self.left) > 0:
            l = self.left[0].depth()
        if len(self.right) > 0:
            r = self.right[0].depth()
        return 1 + max(l, r)


root: Tree = Tree(50)
for v in [30, 70, 20, 40, 60, 80, 10]:
    root.insert(v)
out: list[int] = []
root.walk(out)
print(out, root.depth())

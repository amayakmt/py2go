class Tree:
    def __init__(self, key: int) -> None:
        self.key = key
        self.left = None
        self.right = None

    def insert(self, k: int) -> None:
        if k < self.key:
            if self.left is None:
                self.left = Tree(k)
            else:
                self.left.insert(k)
        else:
            if self.right is None:
                self.right = Tree(k)
            else:
                self.right.insert(k)

    def inorder(self) -> list[int]:
        out: list[int] = []
        if self.left is not None:
            out += self.left.inorder()
        out.append(self.key)
        if self.right is not None:
            out += self.right.inorder()
        return out

    def height(self) -> int:
        lh: int = 0
        rh: int = 0
        if self.left is not None:
            lh = self.left.height()
        if self.right is not None:
            rh = self.right.height()
        return 1 + max(lh, rh)


t: Tree = Tree(50)
for k in [30, 70, 20, 40, 60, 80, 35]:
    t.insert(k)
print(t.inorder(), t.height())

class A:
    x: int = 1

    def show(self):
        return self.x


class B(A):
    pass


class Node:
    def __init__(self, n):
        self.n = n


B.x = 2
Node.empty = Node(0)
Node.top = Node(99)
print(A.x, B.x, A().show(), B().show(), Node.empty.n, Node.top.n)
print(Node(1).top.n)

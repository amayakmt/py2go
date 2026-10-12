from dataclasses import dataclass

type IntList = list[int]
type Pair[T] = tuple[T, T]


class Base[T]:
    def __init__(self, v: T) -> None:
        self.v = v

    def show(self) -> str:
        return "Base:" + str(self.v)


class Child[T](Base[T]):
    def show(self) -> str:
        return "Child:" + str(self.v)


class IntChild(Base[int]):
    def double(self) -> int:
        return self.v * 2


@dataclass
class Node[T]:
    value: T
    children: list["Node[T]"]

    def total(self) -> int:
        return len(self.children)


xs: IntList = [1, 2, 3]
pr: Pair[str] = ("a", "b")
print(xs, pr)
c: Child[str] = Child("z")
print(c.show(), Base(1).show(), IntChild(21).double(), IntChild(2).show())
tree: Node[int] = Node(1, [Node(2, []), Node(3, [])])
print(tree.total(), tree.children[0].value, tree)
items: list[Base[int]] = [Base(1), Child(2), IntChild(3)]
print([i.show() for i in items])


def head[T](n: Node[T]) -> T:
    return n.value


print(head(tree), head(Node("s", [])))

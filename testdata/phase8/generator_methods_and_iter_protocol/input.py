from typing import Iterator


class Tree:
    def __init__(self, v: int, kids: list["Tree"]) -> None:
        self.v = v
        self.kids = kids

    def __iter__(self) -> Iterator[int]:
        yield self.v
        for k in self.kids:
            yield from k


t: Tree = Tree(1, [Tree(2, [Tree(4, [])]), Tree(3, [])])
print(list(t), sum(t), [x * 2 for x in t])
for v in t:
    print(v, end=" ")
print()
it: Iterator[int] = iter([1, 2, 3])
print(next(it), list(it))
it2: Iterator[int] = iter(t)
print(next(it2), next(it2))
words: list[str] = ["a", "b"]
wi: Iterator[str] = iter(words)
print(next(wi), next(wi), next(wi, "end"))
try:
    next(wi)
except StopIteration:
    print("stopped")


class Evens:
    def __init__(self, n: int) -> None:
        self.n = n

    def __iter__(self) -> Iterator[int]:
        for i in range(0, self.n, 2):
            yield i


print(list(Evens(9)), len(list(Evens(4))))

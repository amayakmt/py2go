class Box[T]:
    def __init__(self, item: T) -> None:
        self.item = item

    def get(self) -> T:
        return self.item

    def set(self, item: T) -> None:
        self.item = item

    def __repr__(self) -> str:
        return f"Box({self.item!r})"


class Stack[T]:
    def __init__(self) -> None:
        self.items: list[T] = []

    def push(self, x: T) -> None:
        self.items.append(x)

    def pop(self) -> T:
        return self.items.pop()

    def __len__(self) -> int:
        return len(self.items)


class Pair[K, V]:
    def __init__(self, key: K, value: V) -> None:
        self.key = key
        self.value = value

    def swapped(self) -> "Pair[V, K]":
        return Pair(self.value, self.key)

    def __repr__(self) -> str:
        return f"({self.key!r}: {self.value!r})"


b: Box[int] = Box(5)
s: Box[str] = Box("x")
print(b, s, b.get() + 1, s.get().upper())
b.set(10)
print(b.get(), type(b).__name__, type(s).__name__)
st: Stack[int] = Stack()
st.push(1)
st.push(2)
print(len(st), st.pop(), len(st))
names: Stack[str] = Stack[str]()
names.push("a")
print(names.pop(), len(names))
p: Pair[str, int] = Pair("a", 1)
print(p, p.swapped(), p.swapped().key + 1)
boxes: list[Box[int]] = [Box(1), Box(2)]
print(boxes, sum(x.get() for x in boxes))

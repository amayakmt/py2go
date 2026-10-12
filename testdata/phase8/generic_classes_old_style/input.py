from typing import Generic, TypeVar

T = TypeVar("T")
K = TypeVar("K")
V = TypeVar("V")


class Queue(Generic[T]):
    def __init__(self) -> None:
        self.items: list[T] = []

    def put(self, x: T) -> None:
        self.items.append(x)

    def get(self) -> T:
        return self.items.pop(0)

    def empty(self) -> bool:
        return not self.items


class Registry(Generic[K, V]):
    def __init__(self) -> None:
        self.data: dict[K, V] = {}

    def add(self, k: K, v: V) -> None:
        self.data[k] = v

    def lookup(self, k: K) -> V:
        return self.data[k]


def last(xs: list[T]) -> T:
    return xs[-1]


q: Queue[int] = Queue()
q.put(1)
q.put(2)
print(q.get(), q.empty(), q.get(), q.empty())
r: Registry[str, float] = Registry()
r.add("pi", 3.14)
print(r.lookup("pi"))
print(last([1, 2, 3]), last(["x"]))

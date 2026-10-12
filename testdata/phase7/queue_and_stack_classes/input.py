class Stack:
    def __init__(self) -> None:
        self.data: list[int] = []

    def push(self, x: int) -> None:
        self.data.append(x)

    def pop(self) -> int:
        if not self.data:
            raise IndexError("pop from empty stack")
        return self.data.pop()

    def peek(self) -> int:
        return self.data[-1]

    def __len__(self) -> int:
        return len(self.data)


class Queue:
    def __init__(self) -> None:
        self.inbox = Stack()
        self.outbox = Stack()

    def put(self, x: int) -> None:
        self.inbox.push(x)

    def get(self) -> int:
        if len(self.outbox) == 0:
            while len(self.inbox) > 0:
                self.outbox.push(self.inbox.pop())
        return self.outbox.pop()


q: Queue = Queue()
for i in range(1, 5):
    q.put(i * 10)
print(q.get(), q.get())
q.put(99)
print([q.get() for _ in range(3)])
try:
    q.get()
except IndexError as e:
    print("IndexError:", e)

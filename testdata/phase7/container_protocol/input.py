class Stack:
    def __init__(self) -> None:
        self.items: list[int] = []

    def push(self, v: int) -> None:
        self.items.append(v)

    def __len__(self) -> int:
        return len(self.items)

    def __getitem__(self, i: int) -> int:
        return self.items[i]

    def __setitem__(self, i: int, v: int) -> None:
        self.items[i] = v

    def __delitem__(self, i: int) -> None:
        del self.items[i]

    def __contains__(self, v: int) -> bool:
        return v in self.items

    def __bool__(self) -> bool:
        return len(self.items) > 0


s: Stack = Stack()
print(len(s), bool(s))
s.push(10)
s.push(20)
s.push(30)
print(len(s), bool(s), s[0], s[2], s[-1])
s[1] = 99
print(s[1], 99 in s, 5 in s, 5 not in s)
del s[0]
print(len(s), s[0])
s[0] += 1
print(s[0])
if s:
    print("truthy")

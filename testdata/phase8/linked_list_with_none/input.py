from typing import Optional


class Node:
    def __init__(self, value: int) -> None:
        self.value = value
        self.next: Optional[Node] = None


class LinkedList:
    def __init__(self) -> None:
        self.head: Optional[Node] = None
        self.size = 0

    def push_front(self, v: int) -> None:
        n: Node = Node(v)
        n.next = self.head
        self.head = n
        self.size += 1

    def append(self, v: int) -> None:
        n: Node = Node(v)
        self.size += 1
        if self.head is None:
            self.head = n
            return
        cur: Node = self.head
        while cur.next is not None:
            cur = cur.next
        cur.next = n

    def to_list(self) -> list[int]:
        out: list[int] = []
        cur: Optional[Node] = self.head
        while cur is not None:
            out.append(cur.value)
            cur = cur.next
        return out

    def reverse(self) -> None:
        prev: Optional[Node] = None
        cur: Optional[Node] = self.head
        while cur is not None:
            nxt: Optional[Node] = cur.next
            cur.next = prev
            prev = cur
            cur = nxt
        self.head = prev

    def find(self, v: int) -> Optional[Node]:
        cur: Optional[Node] = self.head
        while cur is not None:
            if cur.value == v:
                return cur
            cur = cur.next
        return None


ll: LinkedList = LinkedList()
for i in range(4):
    ll.append(i)
ll.push_front(-1)
print(ll.to_list(), ll.size)
ll.reverse()
print(ll.to_list())
found: Optional[Node] = ll.find(2)
print(found is not None, ll.find(99) is None)
if found is not None:
    print(found.value, found.next is None)

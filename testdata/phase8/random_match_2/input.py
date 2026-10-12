from typing import Optional


class N:
    def __init__(self, v: int, nxt: Optional['N']) -> None:
        self.v = v
        self.nxt = nxt

def build(xs: list[int]) -> Optional[N]:
    head: Optional[N] = None
    for x in reversed(xs):
        head = N(x, head)
    return head

def total(n: Optional[N]) -> int:
    t: int = 0
    cur: Optional[N] = n
    while cur is not None:
        t += cur.v
        cur = cur.nxt
    return t

def nth(n: Optional[N], k: int) -> Optional[int]:
    cur: Optional[N] = n
    i: int = 0
    while cur is not None:
        if i == k:
            return cur.v
        cur = cur.nxt
        i += 1
    return None

def pick(a: Optional[int], b: Optional[int]) -> int:
    if a is None:
        if b is None:
            return -1
        return b
    if b is None:
        return a
    return a + b


h0: Optional[N] = build([2, 4, 8])
print(total(h0), nth(h0, 2), h0 is None)
h1: Optional[N] = build([7, 8, 2])
print(total(h1), nth(h1, 6), h1 is None)
h2: Optional[N] = build([3, 7])
print(total(h2), nth(h2, 2), h2 is None)
h3: Optional[N] = build([9, 9, 5])
print(total(h3), nth(h3, 5), h3 is None)
h4: Optional[N] = build([1, 9, 0, 6, 2])
print(total(h4), nth(h4, 5), h4 is None)
print(pick(0, 1))
print(pick(1, None))
print(pick(None, 1))
print(pick(5, None))

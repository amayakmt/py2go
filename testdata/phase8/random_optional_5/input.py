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
    return a * b


h0: Optional[N] = build([1, 3, 1])
print(total(h0), nth(h0, 5), h0 is None)
h1: Optional[N] = build([1, 4, 0, 8])
print(total(h1), nth(h1, 2), h1 is None)
h2: Optional[N] = build([9, 5, 5])
print(total(h2), nth(h2, 1), h2 is None)
h3: Optional[N] = build([])
print(total(h3), nth(h3, 4), h3 is None)
print(pick(None, None))
print(pick(4, None))
print(pick(3, None))
print(pick(5, 3))

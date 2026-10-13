from collections import deque
d = deque([1, 2, 3])
d.append(4); d.appendleft(0)
print(d, len(d), d[0], d[-1], list(d), 2 in d, 9 in d, d.count(3), d.index(2), bool(d), bool(deque()), d == deque([0, 1, 2, 3, 4]), d == [0, 1, 2, 3, 4])
print(d.pop(), d.popleft(), d, d.copy(), reversed(d) is not None, list(reversed(d)))
d.extend([7, 8]); d.extendleft([-1, -2])
print(d, d + deque("ab"), d * 2 if False else len(d * 2))
d.rotate(2); print(d)
d.rotate(-3); print(d)
d.insert(1, "x"); d.remove(8); d.reverse()
print(d, d.maxlen)
m = deque(maxlen=3)
for i in range(6):
    m.append(i)
print(m, m.maxlen, len(m))
m.appendleft(99); print(m)
m2 = deque(range(5), 3)
print(m2, deque([1, 2], maxlen=5))
d[0] = "first"; del d[1]
print(d)
q = deque()
for step in range(3):
    q.append(step)
while q:
    print(q.popleft(), end=" ")
print()
for f in (lambda: deque().pop(), lambda: deque().popleft(), lambda: deque([1])[5], lambda: deque([1]).remove(9), lambda: deque([1]).index(9), lambda: deque(maxlen=-1), lambda: deque([1]) + [1]):
    try:
        print(f())
    except (IndexError, ValueError, TypeError) as e:
        print(type(e).__name__, e)

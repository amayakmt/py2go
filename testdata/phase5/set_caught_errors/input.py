s: set[int] = {1}
try:
    s.remove(7)
except KeyError as e:
    print("KeyError", e)
try:
    s.clear()
    s.pop()
except KeyError as e:
    print("KeyError", e)
try:
    t: set[int] = {1, 2}
    for x in t:
        t.discard(x)
except RuntimeError as e:
    print(e)

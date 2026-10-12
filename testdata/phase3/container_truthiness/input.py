empty: list[int] = []
full: list[int] = [0]
t: tuple[int, int] = (0, 0)
print(bool(empty), bool(full), bool(t), bool(""), bool(" "))
print(not empty, not full, not "", not "a")
if empty:
    print("empty is truthy")
if full:
    print("[0] is truthy")
if t:
    print("tuple is truthy")
n: int = 0
while full:
    full.pop()
    n += 1
print(n, len(full))

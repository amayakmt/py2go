stack: list[str] = []
for tok in "( [ { } ] )".split():
    if tok in "([{":
        stack.append(tok)
    else:
        top: str = stack.pop()
        print(top, tok)
print(len(stack) == 0)

queue: list[int] = [1, 2, 3]
while queue:
    item: int = queue.pop(0)
    print(item)
    if item < 3:
        queue.append(item * 10)

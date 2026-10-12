items = [1, "two", 3.0, None, True, [4, 5], (6, 7), {"k": 8}, {9}]
for item in items:
    print(type(item).__name__, repr(item))
print(len(items), items[1], items[-1], items[5][1], items[6][0], items[7]["k"])
items.append("end")
items.insert(0, "start")
print(items.pop(), items.pop(0), len(items))
print([type(x).__name__ for x in items])
print(items.count(1), items.index("two"), 3.0 in items, "nope" in items)
numbers = [x for x in items if isinstance(x, (int, float)) and not isinstance(x, bool)]
print(numbers, sum(numbers))

items = [1, 2]
kind = type(items)
print(kind, type(None), type({}), type((1, "a")), type({1}), type(b"x"))
table = {type(None): "nil", list: "list", dict: "dict"}
print(table[type(None)], table[type(items)], table[type({"k": 1})])
print(type(None) is type(None), type(items) is list, type(items).__name__)

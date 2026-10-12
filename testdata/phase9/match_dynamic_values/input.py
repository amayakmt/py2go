def classify(v):
    match v:
        case None:
            return "none"
        case True | False:
            return "bool"
        case int(n) if n < 0:
            return "negative int"
        case int():
            return "int"
        case float():
            return "float"
        case str() as s if not s:
            return "empty str"
        case str():
            return "str"
        case []:
            return "empty list"
        case [x]:
            return "one item " + classify(x)
        case [x, *rest]:
            return "list starting with " + classify(x) + " +" + str(len(rest))
        case (a, b):
            return "pair"
        case {"kind": kind}:
            return "tagged " + str(kind)
        case {}:
            return "dict"
        case _:
            return "other"

for v in (None, True, -3, 7, 2.5, "", "hi", [], [1], [[2], "x", 3], {"kind": "k"}, {"a": 1}, {1, 2}, (1, 2)):
    print(repr(v), "->", classify(v))

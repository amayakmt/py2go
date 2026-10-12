d: dict[str, int] = {"a": 1}
try:
    print(d["b"])
except KeyError as e:
    print("KeyError", e)
try:
    d.pop("c")
except LookupError as e:
    print(type(e).__name__, e)
try:
    for k in d:
        d["x" + k] = 1
except RuntimeError as e:
    print(e)

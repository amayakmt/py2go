import json

ALPHA = 1
BETA = "two"


def helper():
    return 3


g = globals()
print(sorted(k for k in g if not k.startswith("_") and k.isupper()), g["ALPHA"], g["BETA"], g["helper"](), g["__name__"])
print("json" in g, "helper" in globals())

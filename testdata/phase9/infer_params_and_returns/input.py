def add(a, b):
    return a + b

def describe(x):
    if x is None:
        return "nothing"
    if x:
        return "truthy " + str(x)
    return "falsy " + str(x)

def first(items, default=None):
    for item in items:
        return item
    return default

print(add(1, 2), add(1.5, 2), add("a", "b"), add([1], [2]), add((1,), (2, 3)))
print(describe(None), describe(0), describe("x"), describe([]), describe([0]))
print(first([5, 6]), first([]), first([], "none"), first("abc"))

def apply_twice(f, x):
    return f(f(x))

print(apply_twice(lambda n: n * 2, 5), apply_twice(str.upper, "ab"), apply_twice(len, "abc") if False else 0)

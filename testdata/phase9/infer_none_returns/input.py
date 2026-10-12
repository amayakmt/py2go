def find(items, target):
    for i, item in enumerate(items):
        if item == target:
            return i

def maybe_double(n):
    if n > 0:
        return n * 2

print(find([3, 4, 5], 4), find([3, 4, 5], 9))
r = maybe_double(4)
print(r, maybe_double(-1))
results = [maybe_double(n) for n in (-1, 1, 2)]
print(results, [r for r in results if r is not None])
print(sum(r for r in results if r), find("hello", "l"))

def log(msg):
    print("log:", msg)

value = log("x")
print(value is None, value)

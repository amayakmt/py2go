def apply_all(fs, x):
    return [f(x) for f in fs]

def pipeline(*steps):
    def run(x):
        for step in steps:
            x = step(x)
        return x
    return run

def curry(f, *bound):
    return lambda *rest: f(*bound, *rest)

print(apply_all([abs, str, lambda v: v * 2, float], -3))
print(pipeline(abs, lambda v: v + 1, str, len)(-99))
add3 = curry(lambda a, b, c: a + b + c, 1, 2)
print(add3(10), curry(max, 5)(3, 9), curry(print)("curried"))
ops = {"+": lambda a, b: a + b, "-": lambda a, b: a - b, "max": max, "pow": pow}
for name, op in sorted(ops.items()):
    print(name, op(7, 3))
print(sorted(["bb", "a", "ccc"], key=len), list(map(len, ["bb", "a"])), list(filter(str.isdigit, ["1", "a", "22"])))
print(any(map(str.isupper, "abC")), all(map(bool, [1, "a", [0]])), list(map(int, "123")), sum(map(float, ["1.5", "2"])))
def fact(n):
    return 1 if n <= 1 else n * fact(n - 1)
memo = {}
def fibm(n):
    if n not in memo:
        memo[n] = n if n < 2 else fibm(n - 1) + fibm(n - 2)
    return memo[n]
print(fact(10), fibm(60), list(map(fact, range(6))))

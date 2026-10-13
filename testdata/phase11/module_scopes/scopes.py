x = 10
names = ["a", "b"]
def f(x):
    return x + 1
def g():
    return x
def h():
    x = 5
    def inner():
        return x
    return inner()
def comp():
    return [x for x in range(3)], [n + str(x) for n in names]
def lam():
    return (lambda x: x * 2)(4), (lambda: x)()
class K:
    z = 1
    y = 2
    def m(self):
        return x
def walrus():
    names = [1, 2]
    return names
try:
    raise ValueError("v")
except ValueError as e:
    caught = str(e)
for i in range(3):
    last = i

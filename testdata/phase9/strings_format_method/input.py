print("{} + {} = {}".format(1, 2, 3))
print("{0} {1} {0}".format("a", "b"))
print("{name} is {age:03d}".format(name="Ann", age=7))
print("{:>6}|{:<6}|{:^6}|".format("r", "l", "c"))
print("{!r} {!s}".format("q", "q"))
print("{0[0]} {0[1]}".format([10, 20]), "{p[x]}".format(p={"x": 5}))
class P:
    def __init__(self):
        self.x = 3
        self.y = [4, 5]
print("{0.x} {0.y[1]}".format(P()))
print("{{literal}} {}".format("v"), "{:{w}}".format(5, w=4), "{:.{p}f}".format(3.14159, p=2))
args = ("x", "y")
kwargs = {"k": 1}
print("{} {} {k}".format(*args, **kwargs))
try:
    "{} {}".format(1)
except IndexError as e:
    print("IndexError", e)
try:
    "{x}".format(y=1)
except KeyError as e:
    print("KeyError", e)
print("%s|%5s|%-5s|%05.1f|%x|%c|%%" % ("a", "b", "c", 2.5, 255, 65))
print("%(a)s-%(b)d" % {"a": "x", "b": 3})

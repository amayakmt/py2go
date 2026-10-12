class Greeter:
    def __init__(self, name):
        self.name = name

    def hello(self, greeting="Hello"):
        return greeting + ", " + self.name

    @staticmethod
    def shout(text):
        return text.upper() + "!"

    @classmethod
    def anon(cls):
        return cls("anon")

g = Greeter("Ada")
f = g.hello
print(f(), f("Hi"), f(greeting="Yo"))
m = getattr(g, "hello")
print(m("Hey"), callable(m), callable(g), callable(Greeter))
unbound = Greeter.hello
print(unbound(g), unbound(Greeter("Bob"), "Hola"))
print(Greeter.shout("hey"), g.shout("you"), Greeter.anon().name)
methods = [g.hello, Greeter.shout, Greeter.anon]
print([callable(x) for x in methods], methods[1]("x"))
table = {"greet": g.hello, "up": str.upper, "len": len}
print(table["greet"](), table["up"]("abc"), table["len"]([1, 2]))
maker = Greeter
objs = [maker(n) for n in ("a", "b")]
print([o.name for o in objs], [type(o).__name__ for o in objs])

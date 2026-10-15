import inspect


class A:
    def method(self, x):
        return x


def f(a, b=2, *args, c, d=4, **kw):
    return a


def g():
    pass


print(inspect.isclass(A), inspect.isclass(f), inspect.isclass(3))
print(inspect.isfunction(f), inspect.isfunction(A), inspect.isfunction(len), inspect.isbuiltin(len))
print(inspect.isroutine(f), inspect.isroutine(len), inspect.isroutine(A))
sig = inspect.signature(f)
print(sig)
print(list(sig.parameters))
p = sig.parameters["b"]
print(p.name, p.default, p.kind == inspect.Parameter.POSITIONAL_OR_KEYWORD, sig.parameters["c"].kind == inspect.Parameter.KEYWORD_ONLY)
print(sig.parameters["c"].default is inspect.Parameter.empty)
print(inspect.signature(g))
bound = sig.bind(1, c=3)
bound.apply_defaults()
print(bound.arguments)
try:
    sig.bind(1)
except TypeError as e:
    print("TypeError", e)
print([n for n, v in inspect.getmembers(A) if not n.startswith("__")])
print(inspect.cleandoc("   first\n      second\n    third\n  "))

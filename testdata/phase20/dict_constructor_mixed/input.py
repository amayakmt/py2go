def merged(base, **extra):
    return dict(base, **{}) if not extra else dict(base, flag=True, n=len(extra))


print(dict([("a", 1)], b=2))
print(dict({"x": 1}, y="two", z=[3]))
print(dict(k=1, v="s"), dict(a=1, b=2))
print(merged({"q": 1}), merged({"q": 1}, r=2))
d = {"a": None, "b": 0}
print(d.get("a") is not None, d.get("b") is not None, d.get("c") is not None)
print("%s" % {"k": 1}, "plain" % {"k": 1})
try:
    "plain" % (1, 2)
except TypeError as e:
    print(e)


class NS:
    pass


ns = NS()
ns.zeta = 1
ns.alpha = 2
ns.mid = 3
print(list(vars(ns)))
del ns.alpha
ns.alpha = 9
print(list(vars(ns)), ns.__dict__)

a = [1]
b = a
c = [1]
print(a is b, a is c, a == c, a is not c, None is None, None is not 0, True is True, 1 is 1, "ab" is "ab", () is (), (1,) is not None, id(a) == id(b), id(a) == id(c), id(None) == id(None), id(a) != id(c))
x = None
print(x is None, x is not None, [] is not None, 0 is None, "" is None, False is None, 0 is False, 1 is True, 5 is 5.0 if False else 0, ... is ..., len({id(a), id(b), id(c)}))
class K:
    pass
k1, k2 = K(), K()
k3 = k1
print(k1 is k3, k1 is k2, k1 == k3, k1 == k2, id(k1) == id(k3), id(k1) == id(k2), k1 is not k2, [k1][0] is k1, {1: k1}[1] is k1, (k1, k2)[1] is k2)
def same(p, q):
    return p is q
d = {}
e = d
print(same(d, e), same(d, {}), same(None, None), same(k1, k1))
lst = [k1, k2, k1]
print(lst.index(k2), lst.count(k1), k1 in lst, [x is k1 for x in lst], len({id(x) for x in lst}), lst[0] is lst[2])
t = (a, a)
print(t[0] is t[1], t[0] is a, t == ([1], [1]), t is t)

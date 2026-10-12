class A:
    pass


class B(A):
    pass


class C:
    pass


objs: list[A] = [A(), B()]
a: A = A()
b: B = B()
c: C = C()
print(isinstance(a, (B, C)), isinstance(b, (C, A)), isinstance(c, A), isinstance(c, object))
print(isinstance(1, int), isinstance(True, int), isinstance(1.5, float), isinstance("s", str), isinstance(1, str))
print(isinstance([1], list), isinstance({"a": 1}, dict), isinstance((1, 2), tuple), isinstance(1, (int, str)))
print([isinstance(o, B) for o in objs])

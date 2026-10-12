class C:
    def f(self, value: int) -> int:
        return value


c: C = C()
print(c.f(valu=1))

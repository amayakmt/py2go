class C:
    def f(self, a: int) -> int:
        return a


c: C = C()
print(c.f(1, 2))

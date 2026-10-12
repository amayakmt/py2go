class A:
    def f(self) -> int:
        return 1


a: A = A()
print(A.f(a))

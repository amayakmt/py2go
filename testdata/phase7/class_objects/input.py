class A:
    pass


class B(A):
    def __init__(self) -> None:
        self.me = type(self)


a: A = A()
b: B = B()
print(A, B, type(a), type(b))
print(type(a) is A, type(a) is B, type(b) == B, type(a) == type(b))
print(str(A), repr(B))
print(f"{A} and {type(b)}")
print(A.__name__, type(b).__name__, A.__module__)
print(type(5).__name__, type("s").__name__, type(2.5).__name__, type([1]).__name__)
print(type(5) is int, type("x") == str, type(1.5) is float)
print(b.me)

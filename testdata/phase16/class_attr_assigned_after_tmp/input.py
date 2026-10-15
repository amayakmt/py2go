class A:
    x: int = 1


class B(A):
    pass


B.x = 2

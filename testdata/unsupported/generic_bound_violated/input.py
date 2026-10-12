class A:
    pass


def f[T: A](x: T) -> T:
    return x


f(1)

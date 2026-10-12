class A:
    pass


def make(a: A) -> A:
    return type(a)()

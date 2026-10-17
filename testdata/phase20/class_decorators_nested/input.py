registry = []


def register(cls):
    registry.append(cls.__name__)
    return cls


def tag(name):
    def deco(cls):
        cls.tag = name
        return cls
    return deco


@register
@tag("t1")
class A:
    class Config:
        depth = 3

        def describe(self):
            return "depth=%d" % self.depth

    def cfg(self):
        return self.Config()


class B:
    __class_getitem__ = classmethod(lambda cls, item: cls)

    class Node:
        def __init__(self, v):
            self.v = v

    def make(self, v):
        return B.Node(v)


print(registry, A.tag, A().cfg().describe(), A.Config.depth, type(A().cfg()).__name__)
print(B().make(5).v, B.Node(2).v, type(B.Node(1)).__name__)


def wrapper(cls):
    return 42



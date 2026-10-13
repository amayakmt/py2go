class A:
    def __add__(self, o):
        return self
    __radd__ = __add__

"""abc: abstract base classes (declarations only)."""


class ABC:
    pass


def abstractmethod(funcobj):
    return funcobj


def abstractclassmethod(callable):
    return callable


def abstractstaticmethod(callable):
    return callable


def abstractproperty(callable):
    return callable


def get_cache_token():
    return 0

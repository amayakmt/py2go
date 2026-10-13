"""copy: shallow and deep copies."""


class Error(Exception):
    pass


error = Error


def copy(x):
    return _native("copy.copy", x)


def deepcopy(x, memo=None):
    return _native("copy.deepcopy", x)

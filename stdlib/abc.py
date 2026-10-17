"""Abstract base classes.

@abstractmethod is read by the linker: instantiating a subclass of ABC that
leaves an abstract method unimplemented raises TypeError, as ABCMeta does.
ABC.register(C) makes C a virtual subclass for isinstance and issubclass."""


class ABCMeta:
    """Stands for the metaclass in metaclass=ABCMeta, which the linker turns
    into the checks of ABC."""
    pass


class ABC:

    @classmethod
    def register(cls, subclass):
        _native("abc.register", cls, subclass)
        return subclass


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


def update_abstractmethods(cls):
    return cls

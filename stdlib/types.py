"""types: names for built-in types (the part py2go has)."""


class SimpleNamespace:
    def __init__(self, **kwargs):
        for key, value in kwargs.items():
            setattr(self, key, value)

    def __repr__(self):
        items = []
        for key in self.__dict__:
            items.append(key + "=" + repr(getattr(self, key)))
        return "namespace(" + ", ".join(items) + ")"

    def __eq__(self, other):
        if isinstance(other, SimpleNamespace):
            return self.__dict__ == other.__dict__
        return NotImplemented


class MappingProxyType:
    def __init__(self, mapping):
        self._mapping = mapping

    def __getitem__(self, key):
        return self._mapping[key]

    def __iter__(self):
        return iter(self._mapping)

    def __len__(self):
        return len(self._mapping)

    def __contains__(self, key):
        return key in self._mapping

    def __repr__(self):
        return "mappingproxy(" + repr(self._mapping) + ")"

    def get(self, key, default=None):
        return self._mapping.get(key, default)

    def keys(self):
        return self._mapping.keys()

    def values(self):
        return self._mapping.values()

    def items(self):
        return self._mapping.items()

    def copy(self):
        return self._mapping.copy()


class GenericAlias:
    """C[int] as a value: its origin and arguments. (Classes of the program
    are not subscripted at run time; the linker drops the
    __class_getitem__ = classmethod(GenericAlias) that the library writes.)"""

    def __init__(self, origin, args):
        self.__origin__ = origin
        if not isinstance(args, tuple):
            args = (args,)
        self.__args__ = args

    def __repr__(self):
        def name(a):
            if isinstance(a, type):
                return a.__qualname__
            return repr(a)
        return name(self.__origin__) + "[" + ", ".join(name(a) for a in self.__args__) + "]"

    def __eq__(self, other):
        if isinstance(other, GenericAlias):
            return self.__origin__ == other.__origin__ and self.__args__ == other.__args__
        return NotImplemented

    def __call__(self, *args, **kwargs):
        return self.__origin__(*args, **kwargs)


NoneType = type(None)

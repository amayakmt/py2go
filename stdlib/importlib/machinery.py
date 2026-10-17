"""importlib.machinery: the names of the import machinery (a py2go program
has no finders or loaders of its own; the modules are linked in)."""

SOURCE_SUFFIXES = ['.py']
DEBUG_BYTECODE_SUFFIXES = ['.pyc']
OPTIMIZED_BYTECODE_SUFFIXES = ['.pyc']
BYTECODE_SUFFIXES = ['.pyc']
EXTENSION_SUFFIXES = []


def all_suffixes():
    return SOURCE_SUFFIXES + BYTECODE_SUFFIXES + EXTENSION_SUFFIXES


class ModuleSpec:
    """The import-system-related state of a module."""

    def __init__(self, name, loader, *, origin=None, loader_state=None, is_package=None):
        self.name = name
        self.loader = loader
        self.origin = origin
        self.loader_state = loader_state
        self.submodule_search_locations = [] if is_package else None
        self.cached = None
        self.has_location = False

    @property
    def parent(self):
        if self.submodule_search_locations is None:
            return self.name.rpartition('.')[0]
        return self.name

    def __repr__(self):
        args = [f'name={self.name!r}', f'loader={self.loader!r}']
        if self.origin is not None:
            args.append(f'origin={self.origin!r}')
        return f'ModuleSpec({", ".join(args)})'

    def __eq__(self, other):
        if not isinstance(other, ModuleSpec):
            return NotImplemented
        return self.name == other.name and self.origin == other.origin

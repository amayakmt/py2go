"""grp: the group database."""


class struct_group:
    def __init__(self, fields):
        self.gr_name = fields[0]
        self.gr_passwd = fields[1]
        self.gr_gid = int(fields[2])
        self.gr_mem = [m for m in fields[3].split(",") if m]

    def _t(self):
        return (self.gr_name, self.gr_passwd, self.gr_gid, self.gr_mem)

    def __getitem__(self, i):
        return self._t()[i]

    def __iter__(self):
        return iter(self._t())

    def __len__(self):
        return 4

    def __eq__(self, other):
        return self._t() == tuple(other)

    def __repr__(self):
        return "grp.struct_group(gr_name=%r, gr_passwd=%r, gr_gid=%r, gr_mem=%r)" % self._t()


def getgrall():
    return [struct_group(f) for f in _native("posix.group")]


def getgrnam(name):
    if not isinstance(name, str):
        raise TypeError("getgrnam() argument must be str, not " + type(name).__name__)
    for f in _native("posix.group"):
        if f[0] == name:
            return struct_group(f)
    raise KeyError("getgrnam(): name not found: " + repr(name))


def getgrgid(id):
    if not isinstance(id, int):
        raise TypeError("'" + type(id).__name__ + "' object cannot be interpreted as an integer")
    for f in _native("posix.group"):
        if int(f[2]) == id:
            return struct_group(f)
    raise KeyError("getgrgid(): gid not found: " + str(id))

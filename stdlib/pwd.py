"""pwd: the password database."""


class struct_passwd:
    def __init__(self, fields):
        self.pw_name = fields[0]
        self.pw_passwd = fields[1]
        self.pw_uid = int(fields[2])
        self.pw_gid = int(fields[3])
        self.pw_gecos = fields[4]
        self.pw_dir = fields[5]
        self.pw_shell = fields[6]

    def _t(self):
        return (self.pw_name, self.pw_passwd, self.pw_uid, self.pw_gid, self.pw_gecos, self.pw_dir, self.pw_shell)

    def __getitem__(self, i):
        return self._t()[i]

    def __iter__(self):
        return iter(self._t())

    def __len__(self):
        return 7

    def __eq__(self, other):
        return self._t() == tuple(other)

    def __repr__(self):
        return ("pwd.struct_passwd(pw_name=%r, pw_passwd=%r, pw_uid=%r, pw_gid=%r, pw_gecos=%r, pw_dir=%r, pw_shell=%r)"
                % self._t())


def getpwall():
    return [struct_passwd(f) for f in _native("posix.passwd")]


def getpwnam(name):
    if not isinstance(name, str):
        raise TypeError("getpwnam() argument must be str, not " + type(name).__name__)
    for f in _native("posix.passwd"):
        if f[0] == name:
            return struct_passwd(f)
    raise KeyError("getpwnam(): name not found: " + repr(name))


def getpwuid(uid):
    if not isinstance(uid, int):
        raise TypeError("'" + type(uid).__name__ + "' object cannot be interpreted as an integer")
    for f in _native("posix.passwd"):
        if int(f[2]) == uid:
            return struct_passwd(f)
    raise KeyError("getpwuid(): uid not found: " + str(uid))

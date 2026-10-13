"""struct: interpret bytes as packed binary data."""

from builtins import struct_pm_error as error


def pack(fmt, *args):
    return _native("struct.pack", fmt, *args)


def unpack(fmt, buffer):
    return _native("struct.unpack", fmt, buffer)


def calcsize(fmt):
    return _native("struct.calcsize", fmt)


def unpack_from(fmt, buffer, offset=0):
    return _native("struct.unpack_from", fmt, buffer, offset)


def pack_into(fmt, buffer, offset, *args):
    _native("struct.pack_into", fmt, buffer, offset, *args)


def iter_unpack(fmt, buffer):
    size = calcsize(fmt)
    if size == 0:
        raise error("cannot iteratively unpack with a struct of length 0")
    if len(buffer) % size != 0:
        raise error("iterative unpacking requires a buffer of a multiple of " + str(size) + " bytes")
    out = []
    for i in range(0, len(buffer), size):
        out.append(unpack(fmt, buffer[i:i + size]))
    return iter(out)


def _clearcache():
    pass


class Struct:
    def __init__(self, format):
        if isinstance(format, bytes):
            format = format.decode("ascii")
        self.format = format
        self.size = calcsize(format)

    def pack(self, *args):
        return pack(self.format, *args)

    def unpack(self, buffer):
        return unpack(self.format, buffer)

    def pack_into(self, buffer, offset, *args):
        pack_into(self.format, buffer, offset, *args)

    def unpack_from(self, buffer, offset=0):
        return unpack_from(self.format, buffer, offset)

    def iter_unpack(self, buffer):
        return iter_unpack(self.format, buffer)

    def __repr__(self):
        return "Struct(" + repr(self.format) + ")"

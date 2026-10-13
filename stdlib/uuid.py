"""uuid: UUID objects according to RFC 4122."""

import os
import time
import hashlib

RESERVED_NCS, RFC_4122, RESERVED_MICROSOFT, RESERVED_FUTURE = ["reserved for NCS compatibility", "specified in RFC 4122", "reserved for Microsoft compatibility", "reserved for future definition"]


class SafeUUID:
    safe = 0
    unsafe = -1
    unknown = None


class UUID:
    def __init__(self, hex=None, bytes=None, bytes_le=None, fields=None, int=None, version=None, *, is_safe=None):
        given = 0
        for a in (hex, bytes, bytes_le, fields, int):
            if a is not None:
                given += 1
        if given != 1:
            raise TypeError("one of the hex, bytes, bytes_le, fields, or int arguments must be given")
        value = 0
        if hex is not None:
            h = hex.replace("urn:", "").replace("uuid:", "")
            h = h.strip("{}").replace("-", "")
            if len(h) != 32:
                raise ValueError("badly formed hexadecimal UUID string")
            value = _int_from_hex(h)
        elif bytes_le is not None:
            if len(bytes_le) != 16:
                raise ValueError("bytes_le is not a 16-char string")
            b = bytes_le
            b = b[4 - 1::-1] + b[6 - 1:4 - 1:-1] + b[8 - 1:6 - 1:-1] + b[8:]
            value = _int_from_bytes(b)
        elif bytes is not None:
            if len(bytes) != 16:
                raise ValueError("bytes is not a 16-char string")
            value = _int_from_bytes(bytes)
        elif fields is not None:
            if len(fields) != 6:
                raise ValueError("fields is not a 6-tuple")
            time_low, time_mid, time_hi_version, clock_seq_hi_variant, clock_seq_low, node = fields
            if not 0 <= time_low < 1 << 32:
                raise ValueError("field 1 out of range (need a 32-bit value)")
            if not 0 <= time_mid < 1 << 16:
                raise ValueError("field 2 out of range (need a 16-bit value)")
            if not 0 <= time_hi_version < 1 << 16:
                raise ValueError("field 3 out of range (need a 16-bit value)")
            if not 0 <= clock_seq_hi_variant < 1 << 8:
                raise ValueError("field 4 out of range (need an 8-bit value)")
            if not 0 <= clock_seq_low < 1 << 8:
                raise ValueError("field 5 out of range (need an 8-bit value)")
            if not 0 <= node < 1 << 48:
                raise ValueError("field 6 out of range (need a 48-bit value)")
            clock_seq = (clock_seq_hi_variant << 8) | clock_seq_low
            value = (time_low << 96) | (time_mid << 80) | (time_hi_version << 64) | (clock_seq << 48) | node
        else:
            if not 0 <= int < 1 << 128:
                raise ValueError("int is out of range (need a 128-bit value)")
            value = int
        if version is not None:
            if not 1 <= version <= 5:
                raise ValueError("illegal version number")
            value &= ~(0xc000 << 48)
            value |= 0x8000 << 48
            value &= ~(0xf000 << 64)
            value |= version << 76
        self.int = value
        self.is_safe = is_safe

    def __eq__(self, other):
        if isinstance(other, UUID):
            return self.int == other.int
        return NotImplemented

    def __lt__(self, other):
        if isinstance(other, UUID):
            return self.int < other.int
        return NotImplemented

    def __gt__(self, other):
        if isinstance(other, UUID):
            return self.int > other.int
        return NotImplemented

    def __le__(self, other):
        if isinstance(other, UUID):
            return self.int <= other.int
        return NotImplemented

    def __ge__(self, other):
        if isinstance(other, UUID):
            return self.int >= other.int
        return NotImplemented

    def __hash__(self):
        return hash(self.int)

    def __int__(self):
        return self.int

    def __repr__(self):
        return "UUID(" + repr(str(self)) + ")"

    def __str__(self):
        h = "%032x" % self.int
        return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]

    @property
    def bytes(self):
        return self.int.to_bytes(16, "big")

    @property
    def bytes_le(self):
        b = self.bytes
        return b[4 - 1::-1] + b[6 - 1:4 - 1:-1] + b[8 - 1:6 - 1:-1] + b[8:]

    @property
    def fields(self):
        return (self.time_low, self.time_mid, self.time_hi_version, self.clock_seq_hi_variant, self.clock_seq_low, self.node)

    @property
    def time_low(self):
        return self.int >> 96

    @property
    def time_mid(self):
        return (self.int >> 80) & 0xffff

    @property
    def time_hi_version(self):
        return (self.int >> 64) & 0xffff

    @property
    def clock_seq_hi_variant(self):
        return (self.int >> 56) & 0xff

    @property
    def clock_seq_low(self):
        return (self.int >> 48) & 0xff

    @property
    def time(self):
        return ((self.time_hi_version & 0x0fff) << 48) | (self.time_mid << 32) | self.time_low

    @property
    def clock_seq(self):
        return ((self.clock_seq_hi_variant & 0x3f) << 8) | self.clock_seq_low

    @property
    def node(self):
        return self.int & 0xffffffffffff

    @property
    def hex(self):
        return "%032x" % self.int

    @property
    def urn(self):
        return "urn:uuid:" + str(self)

    @property
    def variant(self):
        if not self.int & (0x8000 << 48):
            return RESERVED_NCS
        elif not self.int & (0x4000 << 48):
            return RFC_4122
        elif not self.int & (0x2000 << 48):
            return RESERVED_MICROSOFT
        else:
            return RESERVED_FUTURE

    @property
    def version(self):
        if self.variant == RFC_4122:
            return (self.int >> 76) & 0xf
        return None


def _int_from_hex(h):
    return int(h, 16)


def _int_from_bytes(b):
    return int.from_bytes(b, "big")


def getnode():
    return int.from_bytes(os.urandom(6), "big") | (1 << 40)


def uuid1(node=None, clock_seq=None):
    nanoseconds = time.time_ns()
    timestamp = nanoseconds // 100 + 0x01b21dd213814000
    if clock_seq is None:
        clock_seq = int.from_bytes(os.urandom(2), "big") & 0x3fff
    time_low = timestamp & 0xffffffff
    time_mid = (timestamp >> 32) & 0xffff
    time_hi_version = (timestamp >> 48) & 0x0fff
    clock_seq_low = clock_seq & 0xff
    clock_seq_hi_variant = (clock_seq >> 8) & 0x3f
    if node is None:
        node = getnode()
    return UUID(fields=(time_low, time_mid, time_hi_version, clock_seq_hi_variant, clock_seq_low, node), version=1)


def uuid3(namespace, name):
    if isinstance(name, str):
        name = bytes(name, "utf-8")
    h = hashlib.md5(namespace.bytes + name).digest()
    return UUID(bytes=h[:16], version=3)


def uuid4():
    return UUID(bytes=os.urandom(16), version=4)


def uuid5(namespace, name):
    if isinstance(name, str):
        name = bytes(name, "utf-8")
    h = hashlib.sha1(namespace.bytes + name).digest()
    return UUID(bytes=h[:16], version=5)


NAMESPACE_DNS = UUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
NAMESPACE_URL = UUID("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
NAMESPACE_OID = UUID("6ba7b812-9dad-11d1-80b4-00c04fd430c8")
NAMESPACE_X500 = UUID("6ba7b814-9dad-11d1-80b4-00c04fd430c8")
NIL = UUID("00000000-0000-0000-0000-000000000000")
MAX = UUID("ffffffff-ffff-ffff-ffff-ffffffffffff")

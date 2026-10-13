"""hmac: keyed-hashing for message authentication (RFC 2104)."""

import hashlib

digest_size = None
trans_5C = bytes((x ^ 0x5C) for x in range(256))
trans_36 = bytes((x ^ 0x36) for x in range(256))


def _make(digestmod, data=b""):
    if isinstance(digestmod, str):
        return hashlib.new(digestmod, data)
    h = digestmod()
    if data:
        h.update(data)
    return h


class HMAC:
    blocksize = 64

    def __init__(self, key, msg=None, digestmod=""):
        if not isinstance(key, (bytes, bytearray)):
            raise TypeError("key: expected bytes or bytearray, but got '" + type(key).__name__ + "'")
        if not digestmod:
            raise TypeError("Missing required argument 'digestmod'.")
        self._digestmod = digestmod
        self._inner = _make(digestmod)
        self._outer = _make(digestmod)
        blocksize = self._inner.block_size
        if len(key) > blocksize:
            key = _make(digestmod, bytes(key)).digest()
        key = bytes(key).ljust(blocksize, b"\0")
        self._outer.update(key.translate(trans_5C))
        self._inner.update(key.translate(trans_36))
        self.digest_size = self._inner.digest_size
        self.block_size = blocksize
        if msg is not None:
            self.update(msg)

    @property
    def name(self):
        return "hmac-" + self._inner.name

    def update(self, msg):
        self._inner.update(msg)

    def copy(self):
        other = HMAC(b"", None, self._digestmod)
        other._inner = self._inner.copy()
        other._outer = self._outer.copy()
        return other

    def _current(self):
        h = self._outer.copy()
        h.update(self._inner.digest())
        return h

    def digest(self):
        return self._current().digest()

    def hexdigest(self):
        return self._current().hexdigest()


def new(key, msg=None, digestmod=""):
    return HMAC(key, msg, digestmod)


def digest(key, msg, digest):
    return HMAC(key, msg, digest).digest()


def compare_digest(a, b):
    if isinstance(a, str) != isinstance(b, str):
        raise TypeError("a bytes-like object is required, not '" + type(a).__name__ + "'")
    if len(a) != len(b):
        return False
    diff = 0
    for x, y in zip(a, b):
        if isinstance(x, str):
            diff |= ord(x) ^ ord(y)
        else:
            diff |= x ^ y
    return diff == 0

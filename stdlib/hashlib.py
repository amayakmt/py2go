"""hashlib: secure hashes."""

algorithms_guaranteed = {"md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha3_224", "sha3_256", "sha3_384", "sha3_512", "shake_128", "shake_256", "blake2b", "blake2s"}
algorithms_available = set(algorithms_guaranteed) | {"sha512_224", "sha512_256"}


def new(name, data=b"", *, usedforsecurity=True):
    n = name.lower().replace("-", "_")
    if n == "blake2b":
        return blake2b(data)
    if n == "blake2s":
        return blake2s(data)
    if n == "shake_128":
        return shake_128(data)
    if n == "shake_256":
        return shake_256(data)
    return _native("hashlib.new", name, data)


def md5(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "md5", data)


def sha1(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha1", data)


def sha224(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha224", data)


def sha256(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha256", data)


def sha384(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha384", data)


def sha512(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha512", data)


def sha3_224(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha3_224", data)


def sha3_256(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha3_256", data)


def sha3_384(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha3_384", data)


def sha3_512(data=b"", *, usedforsecurity=True):
    return _native("hashlib.new", "sha3_512", data)


class _Blake2:
    def __init__(self, wide, data, digest_size, key, salt, person, fanout, depth, leaf_size, node_offset, node_depth, inner_size, last_node):
        self._wide = wide
        self._h = _native("hashlib.blake2", wide, data, digest_size, key, salt, person, fanout, depth, leaf_size, node_offset, node_depth, inner_size, last_node)

    @property
    def name(self):
        return "blake2b" if self._wide else "blake2s"

    @property
    def digest_size(self):
        return self._h.digest_size

    @property
    def block_size(self):
        return self._h.block_size

    def update(self, data):
        self._h.update(data)

    def digest(self):
        return self._h.digest()

    def hexdigest(self):
        return self._h.hexdigest()

    def copy(self):
        if self._wide:
            c = blake2b()
        else:
            c = blake2s()
        c._h = self._h.copy()
        return c

    def __repr__(self):
        return "<" + self.name + " _blake2." + ("blake2b" if self._wide else "blake2s") + " object @ " + hex(id(self)) + ">"


class blake2b(_Blake2):
    SALT_SIZE = 16
    PERSON_SIZE = 16
    MAX_KEY_SIZE = 64
    MAX_DIGEST_SIZE = 64

    def __init__(self, data=b"", /, *, digest_size=64, key=b"", salt=b"", person=b"", fanout=1, depth=1, leaf_size=0, node_offset=0, node_depth=0, inner_size=0, last_node=False, usedforsecurity=True):
        _Blake2.__init__(self, True, data, digest_size, key, salt, person, fanout, depth, leaf_size, node_offset, node_depth, inner_size, last_node)


class blake2s(_Blake2):
    SALT_SIZE = 8
    PERSON_SIZE = 8
    MAX_KEY_SIZE = 32
    MAX_DIGEST_SIZE = 32

    def __init__(self, data=b"", /, *, digest_size=32, key=b"", salt=b"", person=b"", fanout=1, depth=1, leaf_size=0, node_offset=0, node_depth=0, inner_size=0, last_node=False, usedforsecurity=True):
        _Blake2.__init__(self, False, data, digest_size, key, salt, person, fanout, depth, leaf_size, node_offset, node_depth, inner_size, last_node)


def shake_128(data=b"", *, usedforsecurity=True):
    return _native("hashlib.shake", 128, data)


def shake_256(data=b"", *, usedforsecurity=True):
    return _native("hashlib.shake", 256, data)


def pbkdf2_hmac(hash_name, password, salt, iterations, dklen=None):
    return _native("hashlib.pbkdf2_hmac", hash_name, password, salt, iterations, dklen)


def file_digest(fileobj, digest, /, *, _bufsize=2 ** 18):
    if isinstance(digest, str):
        digestobj = new(digest)
    else:
        digestobj = digest()
    while True:
        chunk = fileobj.read(_bufsize)
        if not chunk:
            break
        digestobj.update(chunk)
    return digestobj

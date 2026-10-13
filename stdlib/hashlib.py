"""hashlib: secure hashes."""

algorithms_guaranteed = {"md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha3_224", "sha3_256", "sha3_384", "sha3_512"}
algorithms_available = set(algorithms_guaranteed) | {"sha512_224", "sha512_256"}


def new(name, data=b"", *, usedforsecurity=True):
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

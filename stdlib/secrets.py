"""secrets: random numbers for secrets, from the operating system."""

import random

DEFAULT_ENTROPY = 32

_sysrand = random.SystemRandom()
randbits = _sysrand.getrandbits
choice = _sysrand.choice


def randbelow(exclusive_upper_bound):
    if exclusive_upper_bound <= 0:
        raise ValueError("Upper bound must be positive.")
    return _sysrand._randbelow(exclusive_upper_bound)


def token_bytes(nbytes=None):
    if nbytes is None:
        nbytes = DEFAULT_ENTROPY
    return _native("secrets.token_bytes", nbytes)


def token_hex(nbytes=None):
    return token_bytes(nbytes).hex()


def token_urlsafe(nbytes=None):
    raw = token_bytes(nbytes)
    digits = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
    out = []
    bits = 0
    acc = 0
    for b in raw:
        acc = (acc << 8) | b
        bits += 8
        while bits >= 6:
            bits -= 6
            out.append(digits[(acc >> bits) & 63])
        acc &= (1 << bits) - 1
    if bits:
        out.append(digits[(acc << (6 - bits)) & 63])
    return "".join(out)


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


SystemRandom = random.SystemRandom

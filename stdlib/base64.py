"""base64: Base16, Base32, Base64, Ascii85 and Base85 data encodings."""

import re
import struct
import binascii

bytes_types = (bytes, bytearray)


def _bytes_from_decode_data(s):
    if isinstance(s, str):
        try:
            return s.encode("ascii")
        except UnicodeEncodeError:
            raise ValueError("string argument should contain only ASCII characters")
    if isinstance(s, bytes_types):
        return bytes(s)
    raise TypeError("argument should be a bytes-like object or ASCII string, not '" + type(s).__name__ + "'")


def _check_bytes(s):
    if not isinstance(s, bytes_types):
        raise TypeError("a bytes-like object is required, not '" + type(s).__name__ + "'")
    return bytes(s)


def b64encode(s, altchars=None):
    s = _check_bytes(s)
    encoded = binascii.b2a_base64(s, newline=False)
    if altchars is not None:
        altchars = _check_bytes(altchars)
        if len(altchars) != 2:
            raise AssertionError(repr(len(altchars)))
        return encoded.translate(bytes.maketrans(b"+/", altchars))
    return encoded


def b64decode(s, altchars=None, validate=False):
    s = _bytes_from_decode_data(s)
    if altchars is not None:
        altchars = _bytes_from_decode_data(altchars)
        if len(altchars) != 2:
            raise AssertionError(repr(len(altchars)))
        s = s.translate(bytes.maketrans(altchars, b"+/"))
    return binascii.a2b_base64(s, strict_mode=validate)


def standard_b64encode(s):
    return b64encode(s)


def standard_b64decode(s):
    return b64decode(s)


_urlsafe_encode_translation = bytes.maketrans(b"+/", b"-_")
_urlsafe_decode_translation = bytes.maketrans(b"-_", b"+/")


def urlsafe_b64encode(s):
    return b64encode(s).translate(_urlsafe_encode_translation)


def urlsafe_b64decode(s):
    s = _bytes_from_decode_data(s)
    s = s.translate(_urlsafe_decode_translation)
    return b64decode(s)


_b32alphabet = b"ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
_b32hexalphabet = b"0123456789ABCDEFGHIJKLMNOPQRSTUV"


def _b32encode(alphabet, s):
    s = _check_bytes(s)
    leftover = len(s) % 5
    if leftover:
        s = s + b"\0" * (5 - leftover)
    out = bytearray()
    for i in range(0, len(s), 5):
        c = int.from_bytes(s[i:i + 5], "big")
        for shift in (35, 30, 25, 20, 15, 10, 5, 0):
            out.append(alphabet[(c >> shift) & 31])
    if leftover == 1:
        out[-6:] = b"======"
    elif leftover == 2:
        out[-4:] = b"===="
    elif leftover == 3:
        out[-3:] = b"==="
    elif leftover == 4:
        out[-1:] = b"="
    return bytes(out)


def _b32decode(alphabet, s, casefold=False, map01=None):
    s = _bytes_from_decode_data(s)
    if len(s) % 8:
        raise binascii.Error("Incorrect padding")
    if map01 is not None:
        map01 = _bytes_from_decode_data(map01)
        if len(map01) != 1:
            raise AssertionError(repr(len(map01)))
        s = s.translate(bytes.maketrans(b"01", b"O" + map01))
    if casefold:
        s = s.upper()
    l = len(s)
    s = s.rstrip(b"=")
    padchars = l - len(s)
    decoded = bytearray()
    rev = {}
    for i, c in enumerate(alphabet):
        rev[c] = i
    for i in range(0, len(s), 8):
        quanta = s[i:i + 8]
        acc = 0
        try:
            for c in quanta:
                acc = (acc << 5) + rev[c]
        except KeyError:
            raise binascii.Error("Non-base32 digit found") from None
        acc = acc << (5 * (8 - len(quanta)))
        decoded += acc.to_bytes(5, "big")
    if padchars and decoded:
        lastblocklen = 5 - {6: 4, 4: 3, 3: 2, 1: 1}.get(padchars, 0) if padchars in (6, 4, 3, 1) else -1
        if lastblocklen < 0:
            raise binascii.Error("Incorrect padding")
        decoded[-(5 - lastblocklen):] = b""
        if lastblocklen == 5:
            pass
    return bytes(decoded)


def b32encode(s):
    return _b32encode(_b32alphabet, s)


def b32decode(s, casefold=False, map01=None):
    return _b32decode(_b32alphabet, s, casefold, map01)


def b32hexencode(s):
    return _b32encode(_b32hexalphabet, s)


def b32hexdecode(s, casefold=False):
    return _b32decode(_b32hexalphabet, s, casefold)


def b16encode(s):
    return binascii.hexlify(_check_bytes(s)).upper()


def b16decode(s, casefold=False):
    s = _bytes_from_decode_data(s)
    if casefold:
        s = s.upper()
    if re.search(b"[^0-9A-F]", s):
        raise binascii.Error("Non-base16 digit found")
    return binascii.unhexlify(s)


_a85chars = None
_a85chars2 = None
_A85START = b"<~"
_A85END = b"~>"


def _85encode(b, chars, chars2, pad=False, foldnuls=False, foldspaces=False):
    b = _check_bytes(b)
    padding = (-len(b)) % 4
    if padding:
        b = b + b"\0" * padding
    words = struct.unpack("!%dI" % (len(b) // 4), b)
    chunks = []
    for word in words:
        if foldnuls and not word:
            chunks.append(b"z")
        elif foldspaces and word == 0x20202020:
            chunks.append(b"y")
        else:
            chunk = bytearray(5)
            w = word
            for i in (4, 3, 2, 1, 0):
                w, r = divmod(w, 85)
                chunk[i] = chars[r]
            chunks.append(bytes(chunk))
    if padding and not pad:
        if chunks[-1] == b"z":
            chunks[-1] = chars[0] * 5 if False else bytes(bytearray([chars[0]] * 5))
        chunks[-1] = chunks[-1][:-padding]
    return b"".join(chunks)


def a85encode(b, *, foldspaces=False, wrapcol=0, pad=False, adobe=False):
    chars = bytes(range(33, 118))
    result = _85encode(b, chars, None, pad, True, foldspaces)
    if adobe:
        result = _A85START + result
    if wrapcol:
        wrapcol = max(2 if adobe else 1, wrapcol)
        chunks = [result[i:i + wrapcol] for i in range(0, len(result), wrapcol)]
        if adobe:
            if len(chunks[-1]) + 2 > wrapcol:
                chunks.append(b"")
        result = b"\n".join(chunks)
    if adobe:
        result += _A85END
    return result


def a85decode(b, *, foldspaces=False, adobe=False, ignorechars=b" \t\n\r\x0b"):
    b = _bytes_from_decode_data(b)
    if adobe:
        if not b.endswith(_A85END):
            raise ValueError("Ascii85 encoded byte sequences must end with " + repr(_A85END))
        if b.startswith(_A85START):
            b = b[2:-2]
        else:
            b = b[:-2]
    packI = struct.Struct("!I").pack
    decoded = []
    curr = []
    curr_clear = curr.clear
    curr_append = curr.append
    curr_clear()
    for x in b + b"u" * 4:
        if 33 <= x <= 117:
            curr_append(x)
            if len(curr) == 5:
                acc = 0
                for x in curr:
                    acc = 85 * acc + (x - 33)
                try:
                    decoded.append(packI(acc))
                except struct.error:
                    raise ValueError("Ascii85 overflow") from None
                curr_clear()
        elif x == 122:
            if curr:
                raise ValueError("z inside Ascii85 5-tuple")
            decoded.append(b"\0\0\0\0")
        elif foldspaces and x == 121:
            if curr:
                raise ValueError("y inside Ascii85 5-tuple")
            decoded.append(b"\x20\x20\x20\x20")
        elif x in ignorechars:
            continue
        else:
            raise ValueError("Non-Ascii85 digit found: " + chr(x))
    result = b"".join(decoded)
    padding = 4 - len(curr)
    if padding:
        result = result[:-padding]
    return result


_b85alphabet = (b"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" b"abcdefghijklmnopqrstuvwxyz!#$%&()*+-;<=>?@^_`{|}~")


def b85encode(b, pad=False):
    return _85encode(b, _b85alphabet, None, pad)


def b85decode(b):
    b = _bytes_from_decode_data(b)
    rev = {}
    for i, c in enumerate(_b85alphabet):
        rev[c] = i
    padding = (-len(b)) % 5
    b = b + b"~" * padding
    out = []
    packI = struct.Struct("!I").pack
    for i in range(0, len(b), 5):
        chunk = b[i:i + 5]
        acc = 0
        try:
            for c in chunk:
                acc = acc * 85 + rev[c]
        except KeyError:
            raise ValueError("bad base85 character at position " + str(i + len(chunk) - 1 - 0)) from None
        try:
            out.append(packI(acc))
        except struct.error:
            raise ValueError("base85 overflow in hunk starting at byte " + str(i)) from None
    result = b"".join(out)
    if padding:
        result = result[:-padding]
    return result


MAXLINESIZE = 76
MAXBINSIZE = (MAXLINESIZE // 4) * 3


def encodebytes(s):
    s = _check_bytes(s)
    pieces = []
    for i in range(0, len(s), MAXBINSIZE):
        chunk = s[i:i + MAXBINSIZE]
        pieces.append(binascii.b2a_base64(chunk))
    return b"".join(pieces)


def decodebytes(s):
    s = _check_bytes(s)
    return binascii.a2b_base64(s)


def encode(input, output):
    while True:
        s = input.read(MAXBINSIZE)
        if not s:
            break
        while len(s) < MAXBINSIZE:
            ns = input.read(MAXBINSIZE - len(s))
            if not ns:
                break
            s += ns
        line = binascii.b2a_base64(s)
        output.write(line)


def decode(input, output):
    while True:
        line = input.readline()
        if not line:
            break
        s = binascii.a2b_base64(line)
        output.write(s)

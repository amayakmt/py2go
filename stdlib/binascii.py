"""binascii: conversions between binary data and ASCII."""

from builtins import binascii_pm_Error as Error
from builtins import binascii_pm_Incomplete as Incomplete


def hexlify(data, sep=None, bytes_per_sep=1):
    if sep is None:
        return _native("binascii.hexlify", data)
    return _native("binascii.hexlify", data, sep, bytes_per_sep)


def b2a_hex(data, sep=None, bytes_per_sep=1):
    if sep is None:
        return _native("binascii.hexlify", data)
    return _native("binascii.hexlify", data, sep, bytes_per_sep)


def unhexlify(hexstr):
    return _native("binascii.unhexlify", hexstr)


def a2b_hex(hexstr):
    return _native("binascii.unhexlify", hexstr)


def a2b_base64(data, /, *, strict_mode=False):
    return _native("binascii.a2b_base64", data, strict_mode)


def b2a_base64(data, /, *, newline=True):
    return _native("binascii.b2a_base64", data, newline)


def crc32(data, crc=0, /):
    return _native("binascii.crc32", data, crc)


def crc_hqx(data, crc, /):
    return _native("binascii.crc_hqx", data, crc)


def b2a_uu(data, /, *, backtick=False):
    return _native("binascii.b2a_uu", data, backtick)


def a2b_uu(data, /):
    return _native("binascii.a2b_uu", data)


def b2a_qp(data, quotetabs=False, istext=True, header=False):
    return _native("binascii.b2a_qp", data, quotetabs, istext, header)


def a2b_qp(data, header=False):
    return _native("binascii.a2b_qp", data, header)

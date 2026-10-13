"""string: constants and a few helpers."""

whitespace = " \t\n\r\x0b\x0c"
ascii_lowercase = "abcdefghijklmnopqrstuvwxyz"
ascii_uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
ascii_letters = ascii_lowercase + ascii_uppercase
digits = "0123456789"
hexdigits = digits + "abcdef" + "ABCDEF"
octdigits = "01234567"
punctuation = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
printable = digits + ascii_letters + punctuation + whitespace


def capwords(s, sep=None):
    return (sep or " ").join(x.capitalize() for x in s.split(sep))

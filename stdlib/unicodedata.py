"""unicodedata: the Unicode character database (Unicode 15.1.0)."""

unidata_version = "15.1.0"


def category(chr):
    return _native("unicodedata.category", chr)


def bidirectional(chr):
    return _native("unicodedata.bidirectional", chr)


def east_asian_width(chr):
    return _native("unicodedata.east_asian_width", chr)


def combining(chr):
    return _native("unicodedata.combining", chr)


def mirrored(chr):
    return _native("unicodedata.mirrored", chr)


def decomposition(chr):
    return _native("unicodedata.decomposition", chr)


def decimal(chr, *default):
    return _native("unicodedata.decimal", chr, *default)


def digit(chr, *default):
    return _native("unicodedata.digit", chr, *default)


def numeric(chr, *default):
    return _native("unicodedata.numeric", chr, *default)


def name(chr, *default):
    return _native("unicodedata.name", chr, *default)


def lookup(name):
    return _native("unicodedata.lookup", name)


def normalize(form, unistr):
    return _native("unicodedata.normalize", form, unistr)


def is_normalized(form, unistr):
    return _native("unicodedata.is_normalized", form, unistr)

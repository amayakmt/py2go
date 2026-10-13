chars = ["a", "'", '"', "\\", "\n", "\r", "\t", "\x00", "\x07", "\x1b", "\x7f", "\x80", "\x9f", "\xa0", "\xad", "é", "ǆ", "́", "​", " ", "　", "€", "😀", "\U0001f600", "", "￿", "\U000e0001", "\U0010ffff", "中", "ß"]
for c in chars:
    print(repr(c), len(c), ord(c), c.isprintable(), ascii(c) if False else "")
print(repr("it's"), repr('say "hi"'), repr("both ' and \""), repr("tab\there"), repr("é€😀"), repr("á"), repr("\x00\x01"), repr(" "), repr(""))
print([chr(i) for i in (65, 233, 8364, 128512)], [ord(c) for c in "aé€😀"], chr(0x10ffff) == "\U0010ffff")
for bad in (-1, 0x110000, "a", 1.5):
    try:
        print(chr(bad))
    except (ValueError, TypeError, OverflowError) as e:
        print(type(e).__name__, e)
for bad in ("", "ab", 5):
    try:
        print(ord(bad))
    except TypeError as e:
        print(type(e).__name__, e)
print(str(["é", "'"]), str(("a",)), str({"k": "é"}), repr({"é": 1, 2: "€"}), f"{'é'!r} {'é'!s} {'é'!a}" if False else f"{'é'!r} {'é'!s}")

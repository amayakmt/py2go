chars = [chr(c) for c in list(range(0, 40)) + [0x7e, 0x7f, 0x80, 0x9f, 0xa0, 0xa1, 0xad, 0xff, 0x100, 0x2028, 0x2029, 0x200b, 0x200e, 0xfeff, 0xfffd, 0xd7ff, 0xe000, 0xffff, 0x10000, 0x1f600, 0x10ffff]]
for c in chars:
    print(hex(ord(c)), repr(c), ascii(c), c.isprintable(), len(repr(c)))
for s in ["it's", 'say "hi"', "both ' and \"", "back\\slash", "tab\tnl\ncr\r", "\x00\x01", "é́", "a​b", "\U0001F600!"]:
    print(repr(s), str(s) == s, ascii(s), repr(repr(s)), len(repr(s)))
print(repr(""), repr("'"), repr('"'), repr("'\""), repr("\\"), repr("\\'"), ["it's"], {"k'": 'v"'}, ("a",), {"a"}, repr(["\n"]), str(["\n"]))

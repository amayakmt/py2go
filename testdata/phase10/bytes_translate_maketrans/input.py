table = bytes.maketrans(b"abc", b"xyz")
print(len(table), table[97], table[100], b"aabbcc dd".translate(table), b"hello".translate(None, b"l"), b"hello".translate(table, b"h"), b"abc".translate(bytes(range(256))))
rot13 = bytes.maketrans(b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", b"NOPQRSTUVWXYZABCDEFGHIJKLMnopqrstuvwxyzabcdefghijklm")
msg = b"Hello, World!"
print(msg.translate(rot13), msg.translate(rot13).translate(rot13) == msg)
for f in (lambda: bytes.maketrans(b"ab", b"x"), lambda: b"a".translate(b"short"), lambda: b"a".translate("x")):
    try:
        print(f())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
counts = {}
for ch in b"mississippi":
    counts[ch] = counts.get(ch, 0) + 1
print(counts, {chr(k): v for k, v in sorted(counts.items())}, bytes(sorted(counts)))
xor = bytes(c ^ k for c, k in zip(b"secret!!", b"key" * 3))
print(xor, bytes(c ^ k for c, k in zip(xor, b"key" * 3)))

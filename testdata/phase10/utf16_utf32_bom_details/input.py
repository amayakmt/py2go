s = "A€😀"
for enc in ("utf-16", "utf-16-le", "utf-16-be", "utf-32", "utf-32-le", "utf-32-be", "utf-8", "utf-8-sig"):
    e = s.encode(enc)
    print(enc, e.hex(), len(e), e.decode(enc) == s)
print(b"\xff\xfeA\x00".decode("utf-16"), b"\xfe\xff\x00A".decode("utf-16"), b"A\x00".decode("utf-16"), b"\xff\xfe".decode("utf-16") == "", b"\xef\xbb\xbfA".decode("utf-8-sig"), b"A".decode("utf-8-sig"), b"\xef\xbb\xbf\xef\xbb\xbfA".decode("utf-8-sig") == "﻿A")
for bad, enc in ((b"A", "utf-16"), (b"\x00\xd8", "utf-16-le"), (b"\x00\xdc\x00\x00", "utf-16-le"), (b"A\x00\x00", "utf-32-le"), (b"\x00\x00\x11\x00", "utf-32-le")):
    try:
        print(bad.decode(enc))
    except UnicodeDecodeError as e:
        print(enc, e)
print("Aé".encode("utf-16-le"), "😀".encode("utf-16-be").hex(), "😀".encode("utf-16-le").hex(), "𐍈".encode("utf-16-le").hex(), len("𐍈"), len("𐍈".encode("utf-16-le")) // 2)

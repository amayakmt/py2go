import codecs

print(codecs.encode("héllo", "utf-8"), codecs.decode(b"h\xc3\xa9llo", "utf-8"), codecs.encode("abc", "ascii"), codecs.decode(b"\xe9", "latin-1"), codecs.encode("é", "UTF8"), codecs.encode("é", "latin1"))
print(codecs.lookup("UTF8").name, codecs.lookup("latin-1").name, codecs.lookup("utf_16_le").name, codecs.lookup("ASCII").name, codecs.lookup("utf-8-sig").name)
print(codecs.encode(b"abc", "hex"), codecs.decode(b"616263", "hex"), codecs.encode("Hello", "rot13"), codecs.decode("Uryyb", "rot_13"))
for call in (lambda: codecs.lookup("nope"), lambda: codecs.encode("é", "ascii"), lambda: codecs.decode(b"\xff", "utf-8"), lambda: codecs.encode("é", "ascii", "replace"), lambda: codecs.encode("é", "ascii", "xmlcharrefreplace"), lambda: codecs.decode(b"a\xffb", "utf-8", "replace"), lambda: codecs.decode(b"a\xffb", "utf-8", "ignore"), lambda: codecs.lookup_error("zzz")):
    try:
        print(call())
    except (LookupError, UnicodeError) as e:
        print(type(e).__name__, e)
print(codecs.BOM_UTF8, codecs.BOM_UTF16_LE, codecs.BOM_UTF32_BE, codecs.utf_8_encode("aé"), codecs.utf_8_decode(b"a\xc3"), codecs.utf_8_decode(b"a\xc3", "strict", True) if False else "", codecs.utf_8_decode(b"a\xc3\xa9", "strict", True))
print(codecs.latin_1_decode(b"\xe9\xe8"), codecs.ascii_encode("abc"), codecs.escape_encode(b"a\nb\\c\x00"), codecs.unicode_escape_decode(b"\\u00e9\\n"), codecs.unicode_escape_encode("é\n"))
dec = codecs.getincrementaldecoder("utf-8")()
print(repr(dec.decode(b"\xe2\x82")), repr(dec.decode(b"\xac")), repr(dec.decode(b"x", True)))
print(list(codecs.iterencode(["a", "é"], "utf-8")), list(codecs.iterdecode([b"\xe2\x82", b"\xac!"], "utf-8")))
class Rot(codecs.Codec):
    def encode(self, input, errors="strict"):
        return (input[::-1], len(input))
    def decode(self, input, errors="strict"):
        return (input[::-1], len(input))
def search(name):
    if name == "reverse":
        r = Rot()
        return codecs.CodecInfo(r.encode, r.decode, name="reverse")
    return None


codecs.register(search)
print(codecs.encode("abc", "reverse"), codecs.decode("cba", "reverse"), codecs.lookup("reverse").name)
codecs.unregister(search)
try:
    codecs.encode("a", "reverse")
except LookupError as e:
    print("LookupError", e)
with codecs.open("/tmp/py2go_codecs_test.txt", "w", encoding="utf-8") as f:
    f.write("héllo\nwörld\n")
with codecs.open("/tmp/py2go_codecs_test.txt", "r", encoding="utf-8") as f:
    print(f.read().splitlines())
with open("/tmp/py2go_codecs_test.txt", "rb") as f:
    print(f.read())

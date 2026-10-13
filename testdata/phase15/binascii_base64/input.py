import binascii
import base64
import io

tests = [
    lambda: binascii.hexlify(b"\x00\xff", "-"), lambda: binascii.hexlify(b"abcd", ":", 2), lambda: binascii.hexlify(b"abcde", ":", -2), lambda: binascii.unhexlify("zz"), lambda: binascii.unhexlify(b"abc"), lambda: binascii.a2b_hex("0g"),
    lambda: binascii.unhexlify("DEADbeef"), lambda: binascii.a2b_base64(b"YQ"), lambda: binascii.a2b_base64(b"YQ="), lambda: binascii.a2b_base64(b"Y"), lambda: binascii.a2b_base64(b"Y\x00Q=="), lambda: binascii.a2b_base64(b"YQ==x", strict_mode=True),
    lambda: binascii.a2b_base64(b"YWJj\n"), lambda: binascii.a2b_base64(b"=YQ==", strict_mode=True), lambda: binascii.b2a_base64(b"abc"), lambda: binascii.b2a_base64(b"abc", newline=False), lambda: binascii.b2a_base64(b""),
    lambda: binascii.crc32(b"hello"), lambda: binascii.crc32(b"world", 5), lambda: binascii.crc32(b""), lambda: binascii.crc_hqx(b"hello", 0), lambda: binascii.crc_hqx(b"hello", 0xFFFF),
    lambda: binascii.b2a_uu(b"hello world"), lambda: binascii.a2b_uu(b"+:&5L;&\\@=V]R;&0 \n"), lambda: binascii.b2a_uu(b"", backtick=True), lambda: binascii.b2a_uu(b"x" * 46),
    lambda: binascii.b2a_qp(b"a=b \xe9\n"), lambda: binascii.b2a_qp(b"tab\there  \nend ", quotetabs=True), lambda: binascii.b2a_qp(b"x" * 100), lambda: binascii.b2a_qp(b"a b", header=True), lambda: binascii.a2b_qp(b"a=3Db =E9"), lambda: binascii.a2b_qp(b"soft=\nbreak"), lambda: binascii.a2b_qp(b"a_b", header=True),
    lambda: binascii.hexlify("str"), lambda: binascii.a2b_base64("é"), lambda: issubclass(binascii.Error, ValueError), lambda: binascii.Error.__name__,
]
for t in tests:
    try:
        print(repr(t()))
    except Exception as e:
        print(type(e).__name__, e)
data = b"\xfb\xff hello, world!\x00"
for f in (base64.b64encode, base64.urlsafe_b64encode, base64.standard_b64encode, base64.b32encode, base64.b16encode, base64.a85encode, base64.b85encode, base64.b32hexencode, base64.encodebytes):
    print(f.__name__, f(data), f(b""), f(b"a"), f(b"ab"), f(b"abc"), f(b"abcd"), f(b"abcde"))
for f, s in ((base64.b64decode, "YWJj\n"), (base64.b64decode, "YWJ"), (base64.b64decode, "YW*Jj"), (base64.urlsafe_b64decode, "-_8="), (base64.b32decode, "MFRGG==="), (base64.b32decode, "MFRGGZDF"), (base64.b32decode, "MFRG"), (base64.b16decode, "6162"), (base64.b16decode, "6g"), (base64.a85decode, "@:E_WAS,RgBkhF\"D/O92EH6,BF`qtRH$VAH"), (base64.b85decode, "VPa!sWoBn+X=-b1ZEkOHadLBXb#`}nd3nGP"), (base64.b64decode, "YWJj="), (base64.decodebytes, b"YWJj\nZGVm\n")):
    try:
        print(f.__name__, f(s))
    except Exception as e:
        print(f.__name__, type(e).__name__, e)
print(base64.b64decode("YQ==", validate=True), base64.b64encode(b"ab", altchars=b"-_"), base64.encodebytes(b"x" * 60), base64.b32decode("mfrgg===", casefold=True), base64.a85encode(b"hi", foldspaces=True, wrapcol=3, pad=True, adobe=True), base64.b85encode(b"ab", pad=True), base64.a85encode(b"\x00\x00\x00\x00 abc", foldspaces=True), base64.b16decode("abcd", casefold=True))
for call in (lambda: base64.b64decode("YQ", validate=True), lambda: base64.b64encode("str"), lambda: base64.b64decode("é"), lambda: base64.b32decode("A"), lambda: base64.a85decode("z!!!!!!"), lambda: base64.b85decode("{{{{{")):
    try:
        print(call())
    except Exception as e:
        print(type(e).__name__, e)
out = io.BytesIO()
base64.encode(io.BytesIO(b"file contents " * 10), out)
print(out.getvalue())
back = io.BytesIO()
base64.decode(io.BytesIO(out.getvalue()), back)
print(back.getvalue())

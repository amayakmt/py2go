import hashlib
import hmac
import io

data = b"The quick brown fox jumps over the lazy dog"
ctors = {"md5": hashlib.md5, "sha1": hashlib.sha1, "sha224": hashlib.sha224, "sha256": hashlib.sha256, "sha384": hashlib.sha384, "sha512": hashlib.sha512, "sha3_224": hashlib.sha3_224, "sha3_256": hashlib.sha3_256, "sha3_384": hashlib.sha3_384, "sha3_512": hashlib.sha3_512, "blake2b": hashlib.blake2b, "blake2s": hashlib.blake2s}
for name in ["md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha3_224", "sha3_256", "sha3_384", "sha3_512", "blake2b", "blake2s"]:
    h = hashlib.new(name, data)
    print(name, h.hexdigest(), h.digest_size, h.block_size, h.name)
    print(ctors[name](b"").hexdigest())
print(hashlib.shake_128(data).hexdigest(20), hashlib.shake_256(b"").hexdigest(16), hashlib.shake_128(b"abc").digest(4), hashlib.shake_128().block_size, hashlib.shake_256().name)
h = hashlib.sha256()
h.update(b"The quick brown ")
c = h.copy()
h.update(b"fox jumps over the lazy dog")
print(h.hexdigest(), c.hexdigest(), c.digest() == hashlib.sha256(b"The quick brown ").digest())
print(hashlib.blake2b(b"abc", digest_size=20).hexdigest(), hashlib.blake2b(b"abc", key=b"secret").hexdigest(), hashlib.blake2s(b"abc", salt=b"salt", person=b"pers").hexdigest(), hashlib.blake2b(b"abc", digest_size=1).hexdigest())
print(hashlib.blake2b(b"x" * 300).hexdigest(), hashlib.blake2s(b"y" * 130, key=b"k" * 32).hexdigest())
b = hashlib.blake2b()
b.update(b"a" * 128)
b.update(b"b")
cp = b.copy()
cp.update(b"c" * 200)
print(b.hexdigest(), cp.hexdigest(), hashlib.blake2b.MAX_DIGEST_SIZE, hashlib.blake2s.SALT_SIZE)
for call in (lambda: hashlib.blake2b(digest_size=100), lambda: hashlib.blake2b(key=b"k" * 65), lambda: hashlib.blake2s(salt=b"s" * 9), lambda: hashlib.new("nope"), lambda: hashlib.sha256("str"), lambda: hashlib.shake_128().digest()):
    try:
        print(call())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
print(hashlib.pbkdf2_hmac("sha256", b"password", b"salt", 1000).hex(), hashlib.pbkdf2_hmac("sha1", b"pw", b"s", 10, 8).hex(), hashlib.file_digest(io.BytesIO(data), "sha1").hexdigest(), hashlib.file_digest(io.BytesIO(data), hashlib.md5).hexdigest())
print(sorted(hashlib.algorithms_guaranteed))
m = hmac.new(b"key", data, hashlib.sha256)
print(m.hexdigest(), m.name, m.digest_size, m.block_size, hmac.new(b"key", data, "md5").hexdigest(), hmac.digest(b"k" * 100, b"msg", "sha1").hex(), hmac.new(b"k", digestmod="sha512").hexdigest()[:20])
m2 = hmac.new(b"key", digestmod=hashlib.sha1)
m2.update(b"part1")
m3 = m2.copy()
m2.update(b"part2")
print(m2.hexdigest(), m3.hexdigest(), hmac.compare_digest(b"abc", b"abc"), hmac.compare_digest("abc", "abd"), hmac.compare_digest(b"a", b"ab"))
for call in (lambda: hmac.new("str", b"x", "md5"), lambda: hmac.new(b"k"), lambda: hmac.compare_digest("a", b"a")):
    try:
        print(call())
    except TypeError as e:
        print("TypeError", e)

import hashlib
msgs = [b"", b"abc", b"hello world", bytes(range(256)), b"a" * 1000]
for algo in ("md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha3_256", "sha3_512"):
    print(algo, [hashlib.new(algo, m).hexdigest()[:20] for m in msgs], hashlib.new(algo).digest_size, hashlib.new(algo).block_size, hashlib.new(algo).name)
h = hashlib.sha256()
for chunk in (b"ab", b"c", b""):
    h.update(chunk)
print(h.hexdigest(), h.digest() == bytes.fromhex(h.hexdigest()), len(h.digest()), h.copy().hexdigest() == h.hexdigest())
h2 = h.copy(); h2.update(b"more")
print(h.hexdigest() != h2.hexdigest(), hashlib.md5(b"abc").digest().hex(), hashlib.sha1("é".encode()).hexdigest())
print(hashlib.sha256(b"abc").hexdigest() == hashlib.new("SHA256", b"abc").hexdigest(), hashlib.sha512(data=b"x").hexdigest()[:10] if False else "skip", "md5" in hashlib.algorithms_guaranteed)
for bad in (lambda: hashlib.sha256("text"), lambda: hashlib.new("nope"), lambda: hashlib.md5().update("s"), lambda: hashlib.sha256(5)):
    try:
        print(bad())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)

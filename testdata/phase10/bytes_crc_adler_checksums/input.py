def make_table():
    t = []
    for n in range(256):
        c = n
        for _ in range(8):
            c = (c >> 1) ^ 0xEDB88320 if c & 1 else c >> 1
        t.append(c)
    return t

TABLE = make_table()

def crc32(data, crc=0):
    crc ^= 0xFFFFFFFF
    for b in data:
        crc = TABLE[(crc ^ b) & 0xFF] ^ (crc >> 8)
    return crc ^ 0xFFFFFFFF

def adler32(data):
    a, b = 1, 0
    for byte in data:
        a = (a + byte) % 65521
        b = (b + a) % 65521
    return (b << 16) | a

def fnv1a(data, bits=64):
    h = 0xcbf29ce484222325 if bits == 64 else 0x811c9dc5
    prime = 0x100000001b3 if bits == 64 else 0x01000193
    mask = (1 << bits) - 1
    for byte in data:
        h = ((h ^ byte) * prime) & mask
    return h

for s in (b"", b"a", b"abc", b"hello world", b"The quick brown fox jumps over the lazy dog", bytes(range(256)) * 4):
    print(len(s), hex(crc32(s)), hex(adler32(s)), hex(fnv1a(s)), hex(fnv1a(s, 32)))
print(crc32(b"world", crc32(b"hello ")) == crc32(b"hello world"), TABLE[:4], TABLE[255], len(set(TABLE)))
def murmur_mix(h):
    h ^= h >> 33
    h = (h * 0xff51afd7ed558ccd) & 0xFFFFFFFFFFFFFFFF
    h ^= h >> 33
    h = (h * 0xc4ceb9fe1a85ec53) & 0xFFFFFFFFFFFFFFFF
    return h ^ (h >> 33)
print([hex(murmur_mix(i)) for i in range(4)], murmur_mix(2 ** 64 - 1) == murmur_mix(2 ** 64 - 1))

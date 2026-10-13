def u16(b, off):
    return b[off] | (b[off + 1] << 8)

def u32(b, off):
    return int.from_bytes(b[off:off + 4], "little")

def build(records):
    out = bytearray(b"PK\x03\x04")
    out += len(records).to_bytes(2, "little")
    for name, payload in records:
        nb = name.encode("utf-8")
        out += len(nb).to_bytes(1, "big") + nb + len(payload).to_bytes(4, "little") + payload
    out += (sum(out) & 0xFFFFFFFF).to_bytes(4, "big")
    return bytes(out)

def parse(buf):
    if buf[:4] != b"PK\x03\x04":
        raise ValueError("bad magic")
    n = u16(buf, 4)
    pos = 6
    recs = []
    for _ in range(n):
        ln = buf[pos]
        name = buf[pos + 1:pos + 1 + ln].decode("utf-8")
        pos += 1 + ln
        size = u32(buf, pos)
        recs.append((name, buf[pos + 4:pos + 4 + size]))
        pos += 4 + size
    crc = int.from_bytes(buf[pos:pos + 4], "big")
    return recs, crc == sum(buf[:pos]) & 0xFFFFFFFF

data = build([("a.txt", b"hello"), ("日本.bin", bytes(range(10))), ("empty", b""), ("big", b"\xff" * 300)])
print(len(data), data[:12], data[-4:].hex())
recs, ok = parse(data)
print([(n, len(p), p[:4]) for n, p in recs], ok)
bad = bytearray(data)
bad[10] ^= 1
print(parse(bytes(bad))[1])
try:
    parse(b"XXXX")
except ValueError as e:
    print(e)
try:
    parse(data[:20])
except (IndexError, ValueError) as e:
    print(type(e).__name__)

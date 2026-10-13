data = bytes(range(32))
chunks = [data[i:i + 8] for i in range(0, len(data), 8)]
print(chunks[0], chunks[-1].hex(), len(chunks), b"".join(chunks) == data, [sum(c) for c in chunks])
buf = bytearray(16)
for i, c in enumerate(chunks[:2]):
    buf[i * 8:(i + 1) * 8] = c[::-1]
print(buf, buf.hex(" ", 4))
header = (0xCAFEBABE).to_bytes(4, "big") + (7).to_bytes(2, "little") + b"payload"
magic, ver, rest = int.from_bytes(header[:4], "big"), int.from_bytes(header[4:6], "little"), header[6:]
print(hex(magic), ver, rest, rest.decode(), header[-3:], header[:-3][-4:])
def checksum(bs):
    s = 0
    for byte in bs:
        s = (s * 31 + byte) & 0xFFFFFFFF
    return s
print(checksum(b"hello"), checksum(data), hex(checksum(header)))
def varint(n):
    out = bytearray()
    while True:
        b = n & 0x7F
        n >>= 7
        if n:
            out.append(b | 0x80)
        else:
            out.append(b)
            return bytes(out)
print([varint(n).hex() for n in (0, 1, 127, 128, 300, 16384, 2 ** 40)])

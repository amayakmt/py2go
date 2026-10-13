B64 = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

def b64enc(data):
    out = bytearray()
    for i in range(0, len(data), 3):
        chunk = data[i:i + 3]
        n = int.from_bytes(chunk + b"\x00" * (3 - len(chunk)), "big")
        for j in range(4):
            out.append(B64[(n >> (18 - 6 * j)) & 63])
        if len(chunk) < 3:
            out[len(out) - (3 - len(chunk)):] = b"=" * (3 - len(chunk))
    return bytes(out)

def b64dec(text):
    rev = {c: i for i, c in enumerate(B64)}
    out = bytearray()
    for i in range(0, len(text), 4):
        q = text[i:i + 4]
        pad = q.count(b"=")
        n = 0
        for c in q.replace(b"=", b"A"):
            n = (n << 6) | rev[c]
        out += n.to_bytes(3, "big")[:3 - pad]
    return bytes(out)

for s in (b"", b"f", b"fo", b"foo", b"foob", b"fooba", b"foobar", bytes(range(256))):
    e = b64enc(s)
    print(e[:40], len(e), b64dec(e) == s)

def hexdump(data, width=16):
    lines = []
    for off in range(0, len(data), width):
        chunk = data[off:off + width]
        hexpart = " ".join(f"{b:02x}" for b in chunk)
        text = "".join(chr(b) if 32 <= b < 127 else "." for b in chunk)
        lines.append(f"{off:08x}  {hexpart:<{width * 3 - 1}}  |{text}|")
    return "\n".join(lines)

print(hexdump(b"Hello, World!\x00\x01\x02 \xff\xfe and more bytes here"))

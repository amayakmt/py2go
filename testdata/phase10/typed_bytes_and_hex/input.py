def to_hex(b: bytes) -> str:
    return b.hex()

def checksum(data: bytes) -> int:
    total = 0
    for byte in data:
        total = (total + byte) % 256
    return total

def xor_bytes(a: bytes, b: bytes) -> bytes:
    return bytes(x ^ y for x, y in zip(a, b))

data: bytes = b"\x00\x01\xfe\xffhello"
print(to_hex(data), checksum(data), xor_bytes(b"abc", b"\x01\x02\x03"), data[2:6], data[-5:].decode(), len(data), data.find(b"hello"), data.count(b"l"), data.upper(), data[4:] + b"!", data * 2, bytes(reversed(data[:3])), list(data[:4]), max(data), sum(data), data.startswith(b"\x00"), b"ell" in data)
buf: bytearray = bytearray(b"abc")
buf.extend(b"def")
buf[0] = ord("A")
buf.append(33)
buf += b"?"
print(buf, len(buf), bytes(buf), buf.decode(), buf.hex(), buf.pop(), buf, buf.index(b"d"), buf[1:3], buf == b"Abcdef!", sorted(buf)[:3], bytes(sorted(buf)))
words: list[bytes] = [b"pear", b"fig", b"apple"]
print(sorted(words), sorted(words, key=len), b",".join(words), [w.decode().title() for w in words], {w: len(w) for w in words}, b"fig" in words, words.index(b"apple"), max(words), min(words, key=len))

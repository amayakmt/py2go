parts = []
for i in range(2000):
    parts.append(str(i * i))
s = ",".join(parts)
print(len(s), s[:30], s[-20:], s.count(","), hash(s) == hash(",".join(parts)), s.encode()[:10], sum(map(int, s.split(","))))
acc = ""
for i in range(300):
    acc += chr(0x41 + i % 26) + chr(0x3b1 + i % 24)
print(len(acc), acc[:20], acc.upper() == acc.casefold().upper(), len(acc.encode("utf-8")), acc.count("α"), acc.find("Zρ"), sorted(set(acc))[:5], len(set(acc)))
b = bytearray()
for i in range(1000):
    b += (i * 31 % 251).to_bytes(1, "big")
    b.extend(b"\xff" if i % 100 == 0 else b"")
print(len(b), b[:10].hex(), sum(b), max(b), b.count(255), bytes(b).find(b"\xff\xff"), hash(bytes(b)) == hash(bytes(b)), bytes(b)[500:510])
big = 1
for i in range(1, 400):
    big = big * i + i
print(len(str(big)), str(big)[:25], big % 10 ** 18, big.bit_length(), hex(big)[:20], len(bin(big)))
lines = [f"{i:05d}:{'x' * (i % 7)}:{i ** 3:>12}:{i / 7:8.3f}" for i in range(0, 50, 7)]
print("\n".join(lines))

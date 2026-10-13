class Packet:
    def __init__(self, kind, payload):
        self.kind, self.payload = kind, payload
    def __bytes__(self):
        return bytes([self.kind]) + len(self.payload).to_bytes(2, "big") + self.payload
    def __str__(self):
        return f"<Packet {self.kind} {self.payload!r}>"
    def __repr__(self):
        return f"Packet({self.kind!r}, {self.payload!r})"
    def __len__(self):
        return len(self.payload)
    def __iter__(self):
        return iter(self.payload)
    def __contains__(self, b):
        return b in self.payload
    def __getitem__(self, i):
        return self.payload[i]
    def __eq__(self, o):
        return bytes(self) == bytes(o)
    def __hash__(self):
        return hash(bytes(self))

p = Packet(7, b"hello")
print(bytes(p), bytes(p).hex(), str(p), repr(p), len(p), list(p), 104 in p, 1 in p, p[0], p[1:3], p[-1], b"".join([bytes(p), bytes(p)]), len(bytes(Packet(1, b""))))
print(p == Packet(7, b"hello"), p != Packet(7, b"hellx"), sorted(p), max(p), sum(p), bytes(list(p)), bytearray(bytes(p)), f"{p}", "%s|%r" % (p, p), [str(x) for x in (p, Packet(0, b"\x00"))])
print(bytes(p)[3:].decode(), bytes(p)[:1], int.from_bytes(bytes(p)[1:3], "big"), bytes(bytearray(b"ab")))

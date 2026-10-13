import uuid

u = uuid.UUID("12345678-1234-5678-1234-567812345678")
print(u, repr(u), u.hex, u.int, u.bytes, u.bytes_le, u.fields, u.version, u.variant, u.urn, u.time_low, u.time_mid, u.time_hi_version, u.clock_seq, u.node, u.time, hash(u) == hash(uuid.UUID(int=u.int)))
print(uuid.UUID("{12345678-1234-5678-1234-567812345678}") == u, uuid.UUID("urn:uuid:12345678-1234-5678-1234-567812345678") == u, uuid.UUID(hex="12345678123456781234567812345678") == u, uuid.UUID(bytes=u.bytes) == u, uuid.UUID(bytes_le=u.bytes_le) == u, uuid.UUID(fields=u.fields) == u, u < uuid.UUID(int=u.int + 1), u != uuid.UUID(int=0), str(uuid.UUID(int=0)))
print(uuid.uuid3(uuid.NAMESPACE_DNS, "python.org"), uuid.uuid5(uuid.NAMESPACE_DNS, "python.org"), uuid.uuid3(uuid.NAMESPACE_URL, "http://x"), uuid.uuid5(uuid.NAMESPACE_OID, "1.2"), uuid.uuid5(uuid.NAMESPACE_X500, b"cn"))
a, b = uuid.uuid4(), uuid.uuid4()
print(a != b, a.version, a.variant, len(str(a)), uuid.uuid1().version, uuid.uuid1(node=0x123456789abc, clock_seq=5).node, uuid.uuid1(node=0x123456789abc, clock_seq=5).clock_seq, uuid.UUID(int=5, version=4), uuid.NAMESPACE_URL)
for call in (lambda: uuid.UUID("abc"), lambda: uuid.UUID(), lambda: uuid.UUID(hex="a", int=1), lambda: uuid.UUID(bytes=b"abc"), lambda: uuid.UUID(int=1 << 128), lambda: uuid.UUID(int=1, version=9), lambda: uuid.UUID(fields=(1, 2, 3)), lambda: uuid.UUID("zzzzzzzz-1234-5678-1234-567812345678")):
    try:
        print(call())
    except (ValueError, TypeError) as e:
        print(type(e).__name__, e)
print(sorted([uuid.UUID(int=3), uuid.UUID(int=1), uuid.UUID(int=2)])[0].int, {u: 1}[uuid.UUID(str(u))], int(u) == u.int, uuid.UUID(int=0xC000 << 48).variant, uuid.UUID(int=0x4000 << 48).variant, uuid.UUID(int=0x2000 << 48).variant)

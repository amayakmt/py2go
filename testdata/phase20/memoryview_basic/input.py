ba = bytearray(b"hello world")
mv = memoryview(ba)
print(len(mv), mv[0], mv[-1], bytes(mv[0:5]), mv[6:].tobytes(), mv.nbytes, mv.readonly, mv.format, mv.itemsize)
mv[0] = ord("H")
mv[6:11] = b"WORLD"
print(ba, mv.tolist()[:3], mv.hex()[:6], list(mv[:2]))
ro = memoryview(b"abc")
print(ro.readonly, ro == b"abc", bytes(ro[::2]), ro.obj)
try:
    ro[0] = 1
except TypeError as e:
    print("TypeError", e)
with memoryview(b"xyz") as m2:
    print(m2.tobytes())
try:
    m2.tobytes()
except ValueError as e:
    print("ValueError", e)
try:
    memoryview("text")
except TypeError as e:
    print("TypeError", e)
print(b"".join([memoryview(b"ab"), b"cd"]), bytearray(memoryview(b"q")))

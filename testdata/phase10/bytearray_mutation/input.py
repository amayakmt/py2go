a = bytearray(b"hello")
print(a, len(a), a[0], a[1:3], type(a[1:3]).__name__, repr(a), str(a), a == b"hello", a == bytearray(b"hello"), a < b"help", bytearray(), bytearray(3), bytearray([1, 2]), bytearray("hi", "ascii"), bool(bytearray()))
a[0] = 72
a[1:3] = b"EY"
a.append(33)
a.extend(b"??")
a.extend([65, 66])
a += b"+"
a *= 2
print(a, len(a))
a.insert(0, 90)
a.insert(100, 89)
a.insert(-1, 88)
print(a, a.pop(), a.pop(0), a.pop(-2), a)
a.remove(63)
a.reverse()
print(a, a.copy() == a, a.copy() is a, a.index(72), a.count(65))
del a[0]
del a[2:5]
del a[::2]
print(a)
a.clear()
print(a, len(a))
b = bytearray(b"abcdef")
b[1:4] = b""
b[1:1] = b"XYZ"
b[::2] = b"123"
print(b)
b[2:] = bytearray(b"!")
print(b, b.upper(), b.decode(), bytes(b), bytearray(bytes(b)) == b)
def setitem(k, v):
    b[k] = v

for f in (lambda: setitem(0, 256), lambda: setitem(0, -1), lambda: setitem(0, "a"), lambda: b.pop(100), lambda: bytearray().pop(), lambda: b.remove(200), lambda: b.append(300), lambda: b.extend(["a"]), lambda: setitem(slice(0, 4, 2), b"toolong"), lambda: hash(b)):
    try:
        print(f())
    except (ValueError, TypeError, IndexError) as e:
        print(type(e).__name__, e)

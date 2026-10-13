import array
import io

tests = [("b", 200), ("b", -200), ("B", 300), ("B", -1), ("h", 40000), ("H", 70000), ("H", -1), ("i", 2**31), ("I", -1), ("I", 2**32), ("l", 2**63), ("q", 2**63), ("Q", -1), ("Q", 2**64), ("b", 1.5), ("b", "a"), ("d", "a"), ("f", None), ("i", True), ("d", 3), ("f", 0.1)]
for tc, v in tests:
    try:
        print(tc, v, array.array(tc, [v]))
    except Exception as e:
        print(tc, repr(v), type(e).__name__, e)
a = array.array("i", [1, 2, 3])
cases = [lambda: a.pop(10), lambda: a.remove(9), lambda: a.index(9), lambda: array.array("z"), lambda: array.array("i", "abc"), lambda: array.array("i", b"abc"), lambda: a.frombytes(b"abc"), lambda: a.insert(0, "x"), lambda: a + array.array("d"), lambda: a * 2.5, lambda: a[10], lambda: array.array("i", [1]).__setitem__(5, 1), lambda: array.array("i", [1, 2]).tounicode(), lambda: array.array("i", ["x"])]
for f in cases:
    try:
        print(f())
    except Exception as e:
        print(type(e).__name__, e)
print(array.array("d", [1.5, 2]), array.array("f", [0.1]), repr(array.array("B", b"abc")), array.array("i").itemsize, array.typecodes, array.array("i"), array.array("l", [1]).itemsize, array.array("h", [-1]).tobytes())
print(a.buffer_info()[1], a.tobytes(), a.tolist(), a[1:], a[::-1], a * 2, a + a, a == array.array("i", [1, 2, 3]), a < array.array("i", [1, 2, 4]), list(a), 2 in a, len(a), a.typecode, a.itemsize, a[-1])
a.byteswap()
print(a, a.count(16777216))
a.byteswap()
a.reverse(); a.append(7); a.extend([8, 9]); a.insert(1, 100)
print(a, a.pop(), a.pop(0), a, a.index(2), a.count(1))
del a[0]
a[0:1] = array.array("i", [5, 5])
a[1] = 6
a += array.array("i", [4])
a *= 2
print(a, sum(a), max(a), sorted(a), [x * 2 for x in a], 3 * array.array("b", [1]))
buf = io.BytesIO()
array.array("H", [1, 2, 3, 65535]).tofile(buf)
print(buf.getvalue())
buf.seek(0)
b = array.array("H")
b.fromfile(buf, 4)
print(b, b.tobytes() == buf.getvalue())
b.fromlist([7, 8])
b.frombytes(b"\x09\x00")
print(b, array.array("B", range(5)), array.array("B", array.array("B", [1, 2])), array.array("d", array.array("i", [1, 2])))
print(array.array("u", "hello"), array.array("w", "wörld").tounicode(), array.array("u", "ab").tobytes())

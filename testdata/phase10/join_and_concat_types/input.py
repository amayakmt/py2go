print(",".join(["a", "b"]), ",".join(("a", "b")), ",".join({"k": 1}), ",".join("abc"), ",".join(map(str, range(3))), ",".join([]), ",".join(["x"]), "".join(reversed("abc")), "é".join(["日", "本"]), b",".join([b"a", bytearray(b"b")]), bytearray(b",").join([b"a", b"b"]), "-".join(str(i) for i in range(4)))
for f in (lambda: ",".join([1]), lambda: ",".join([b"a"]), lambda: b",".join(["a"]), lambda: ",".join(5), lambda: ",".join(None), lambda: ",".join(["a", None]), lambda: b",".join([1, 2]), lambda: ",".join([["a"]])):
    try:
        print(f())
    except TypeError as e:
        print(e)
print("a" + "b", "a" "b", b"a" + b"b", b"a" + bytearray(b"b"), bytearray(b"a") + b"b", [1] + [2], (1,) + (2,), "x" * 2 + "y", str(1) + str(2.5) + str(None) + str(True) + str(2 ** 70) + str(1j) + str(b"b") + str([1]) + str({1}) + str((1,)))
for f in (lambda: "a" + 1, lambda: 1 + "a", lambda: "a" + None, lambda: "a" + b"b", lambda: b"a" + "b", lambda: [1] + (1,), lambda: (1,) + [1], lambda: "a" + ["b"], lambda: bytearray(b"a") + "b"):
    try:
        print(f())
    except TypeError as e:
        print(e)

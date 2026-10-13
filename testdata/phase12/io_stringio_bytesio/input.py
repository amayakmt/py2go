import io
s = io.StringIO()
print(s.write("hello "), s.write("world\nline2\n"), s.tell(), repr(s.getvalue()))
s.seek(0)
print(repr(s.readline()), repr(s.read(3)), repr(s.read()), repr(s.read()), s.readline() == "")
s.seek(6); s.write("WORLD")
print(repr(s.getvalue()), s.tell())
s.truncate(5); print(repr(s.getvalue()))
t = io.StringIO("a\nb\nc")
print(list(t), t.read() == "", io.StringIO("x\ny\n").readlines(), [l for l in io.StringIO("1\n2\n")])
with io.StringIO("ctx\n") as f:
    print(f.read(), f.closed)
print(f.closed)
try:
    f.read()
except ValueError as e:
    print(e)
try:
    io.StringIO().write(5)
except TypeError as e:
    print(e)
b = io.BytesIO(b"bytes data")
print(b.read(5), b.tell(), b.read(), b.getvalue(), b.seek(0), b.readline(), b.seek(-4, 2), b.read())
b2 = io.BytesIO()
print(b2.write(b"ab"), b2.write(bytearray(b"cd")), b2.getvalue(), b2.seek(1), b2.write(b"X"), b2.getvalue(), list(io.BytesIO(b"l1\nl2")))
try:
    io.BytesIO().write("str")
except TypeError as e:
    print(e)
out = io.StringIO()
print("to buffer", 42, file=out)
print("second", end="", file=out)
print(repr(out.getvalue()))

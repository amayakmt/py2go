import io

class Upper(io.TextIOBase):
    def __init__(self):
        io.TextIOBase.__init__(self)
        self.parts = []

    def writable(self):
        return True

    def write(self, s):
        self.parts.append(s.upper())
        return len(s)

u = Upper()
print(u.write("abc"), u.writable(), u.readable(), u.closed, u.encoding, u.parts)
u.close()
print(u.closed)
try:
    u.flush()
except ValueError as e:
    print(e)

class Counter(io.RawIOBase):
    def __init__(self, n):
        io.RawIOBase.__init__(self)
        self.n = n

    def readable(self):
        return True

    def readinto(self, b):
        k = min(len(b), self.n)
        for i in range(k):
            b[i] = 65 + i % 26
        self.n -= k
        return k

c = Counter(30)
print(c.read(5), c.read(100), c.read(3), c.readable(), c.seekable())
print(Counter(10).readall())

raw = io.BytesIO(b"line one\nline two\r\nline three")
tw = io.TextIOWrapper(raw, encoding="utf-8")
print(tw.readline(), repr(tw.readline()), tw.read(4), tw.encoding, tw.errors, tw.closed, tw.readable())
tw.seek(0)
print(tw.readlines(), tw.newlines)
raw = io.BytesIO()
tw = io.TextIOWrapper(raw, encoding="latin-1", newline="\r\n", line_buffering=True)
print(tw.write("a\nb\n"), raw.getvalue(), tw.line_buffering)
tw.write("é")
tw.flush()
print(raw.getvalue())
b = tw.detach()
print(b is raw)
tw = io.TextIOWrapper(io.BytesIO("héllo wörld\n".encode("utf-8")))
print(list(tw), tw.tell() >= 0)
with io.TextIOWrapper(io.BytesIO(b"a\nb\n"), newline=None) as t:
    print([line for line in t])
print(t.closed)

br = io.BufferedReader(io.BytesIO(b"hello world\nsecond line\n"))
print(br.read(5), br.peek(3)[:3], br.readline(), br.read1(3), br.read(), br.read())
bw = io.BufferedWriter(io.BytesIO())
bw.write(b"abc")
bw.write(b"def")
raw = bw.raw
print(raw.getvalue(), bw.tell())
bw.flush()
print(raw.getvalue())
f = io.FileIO("/dev/null", "r")
print(f.read(), f.closed)
f.close()
print(f.closed)
print(io.DEFAULT_BUFFER_SIZE, io.SEEK_END, io.text_encoding(None), issubclass(io.UnsupportedOperation, OSError))
try:
    io.TextIOWrapper(io.BytesIO(), newline="x")
except ValueError as e:
    print(e)

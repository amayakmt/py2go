import gzip
import io
import os

data = b"line one\nline two\nline three\n" * 20
c = gzip.compress(data, mtime=0)
print(c[:10], gzip.decompress(c) == data, len(gzip.decompress(c)), gzip.decompress(c + gzip.compress(b"second", mtime=0)).endswith(b"second"))
print(gzip.compress(b"", mtime=0)[:10], gzip.decompress(gzip.compress(b"")), gzip.compress(b"x", 1, mtime=5)[:10], gzip.compress(b"x", 6, mtime=5)[:10], gzip.compress(b"x", mtime=1234567)[4:8])
path = "/tmp/py2go_gzip_test.gz"
with gzip.open(path, "wb") as f:
    print(f.write(b"hello\nworld\n"), f.mode, f.writable(), f.readable())
    f.write(b"more")
with gzip.open(path, "rb") as f:
    print(f.read(5), f.readline(), f.readline(), f.read(), f.read(), f.mode)
with gzip.open(path, "rb") as f:
    print(list(f), f.tell() if False else "")
with gzip.open(path) as f:
    f.seek(6)
    print(f.read(5), f.tell())
with gzip.open(path, "wt", encoding="utf-8") as f:
    f.write("héllo\n")
    f.write("wörld\n")
with gzip.open(path, "rt", encoding="utf-8") as f:
    print(f.readline(), f.read())
with gzip.open(path, "rt") as f:
    print(list(f))
with open(path, "rb") as raw:
    head = raw.read(10)
    print(head[:3], head[3], head[8:], os.path.getsize(path) > 18)
buf = io.BytesIO()
with gzip.GzipFile(fileobj=buf, mode="wb", mtime=0) as g:
    g.write(b"in memory")
print(buf.getvalue()[:10], gzip.decompress(buf.getvalue()))
buf.seek(0)
print(gzip.GzipFile(fileobj=buf).read(), gzip.GzipFile(fileobj=io.BytesIO(buf.getvalue())).mtime if False else "")
for call in (lambda: gzip.decompress(b"not gzip"), lambda: gzip.decompress(c[:-4]), lambda: gzip.decompress(c[:-8] + b"\x00" * 8), lambda: gzip.open(path, "rt", encoding=None, newline=None) and None, lambda: gzip.open(path, "rb", encoding="utf-8"), lambda: gzip.open(path, "rbt"), lambda: gzip.open(5), lambda: gzip.GzipFile(fileobj=io.BytesIO(b"junk")).read()):
    try:
        print(call())
    except (gzip.BadGzipFile, EOFError, ValueError, TypeError, OSError) as e:
        print(type(e).__name__, e)
print(gzip.BadGzipFile.__mro__[1].__name__ if False else issubclass(gzip.BadGzipFile, OSError), gzip.READ, gzip.WRITE)
os.remove(path)

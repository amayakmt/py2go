import io
import os
import tarfile
import tempfile

buf = io.BytesIO()
with tarfile.open(fileobj=buf, mode="w") as tf:
    ti = tarfile.TarInfo("hello.txt")
    ti.size = 6
    ti.mtime = 1600000000
    ti.mode = 0o644
    ti.uname = "alice"
    ti.gname = "staff"
    tf.addfile(ti, io.BytesIO(b"hello\n"))
    d = tarfile.TarInfo("dir")
    d.type = tarfile.DIRTYPE
    d.mode = 0o755
    d.mtime = 1600000000
    tf.addfile(d)
    long = tarfile.TarInfo("dir/" + "x" * 120 + ".txt")
    long.size = 3
    long.mtime = 1600000000
    tf.addfile(long, io.BytesIO(b"abc"))
    uni = tarfile.TarInfo("dir/ünï.txt")
    uni.size = 1
    uni.mtime = 1600000000.5
    tf.addfile(uni, io.BytesIO(b"u"))
    ln = tarfile.TarInfo("link")
    ln.type = tarfile.SYMTYPE
    ln.linkname = "hello.txt"
    ln.mtime = 1600000000
    tf.addfile(ln)
data = buf.getvalue()
print(len(data), len(data) % 10240, data[:5], data[257:263], data[100:108], data[148:156], data[512 * 0:512 * 0 + 6])
print(tarfile.is_tarfile(io.BytesIO(data)), tarfile.is_tarfile(io.BytesIO(b"x" * 1000)))
with tarfile.open(fileobj=io.BytesIO(data)) as tf:
    print(tf.getnames())
    for m in tf:
        print(m.name, m.size, m.type, oct(m.mode), m.mtime, m.uname, m.gname, m.isfile(), m.isdir(), m.issym(), m.linkname)
    print(tf.extractfile("hello.txt").read(), tf.extractfile("dir") is None, tf.getmember("dir").isdir())
    try:
        tf.getmember("nope")
    except KeyError as e:
        print("KeyError", e)
for fmt in (tarfile.USTAR_FORMAT, tarfile.GNU_FORMAT, tarfile.PAX_FORMAT):
    b = io.BytesIO()
    with tarfile.open(fileobj=b, mode="w", format=fmt) as tf:
        ti = tarfile.TarInfo("a/b.txt")
        ti.size = 2
        ti.mtime = 5
        tf.addfile(ti, io.BytesIO(b"hi"))
        ti = tarfile.TarInfo("p/" + "q" * 150)
        ti.mtime = 5
        try:
            tf.addfile(ti)
        except ValueError as e:
            print(fmt, "ValueError", e)
    out = b.getvalue()
    print(fmt, len(out), out[257:265], tarfile.open(fileobj=io.BytesIO(out)).getnames())

d = tempfile.mkdtemp()
src = os.path.join(d, "src")
os.makedirs(os.path.join(src, "sub"))
with open(os.path.join(src, "f.txt"), "w") as f:
    f.write("file contents\n" * 50)
with open(os.path.join(src, "sub", "g.bin"), "wb") as f:
    f.write(bytes(range(256)))
for mode, name in (("w", "plain.tar"), ("w:gz", "z.tar.gz")):
    path = os.path.join(d, name)
    with tarfile.open(path, mode) as tf:
        tf.add(src, arcname="root")
    with tarfile.open(path) as tf:
        print(name, tf.getnames())
        tf.extractall(os.path.join(d, "out_" + name), filter="data")
    print(sorted(os.listdir(os.path.join(d, "out_" + name, "root"))), open(os.path.join(d, "out_" + name, "root", "sub", "g.bin"), "rb").read() == bytes(range(256)))
print(open(os.path.join(d, "z.tar.gz"), "rb").read(2))
with tarfile.open(os.path.join(d, "plain.tar"), "a") as tf:
    tf.add(os.path.join(src, "f.txt"), arcname="appended.txt")
with tarfile.open(os.path.join(d, "plain.tar")) as tf:
    print(tf.getnames())
    tf.list(verbose=False)
try:
    tarfile.open(os.path.join(d, "f.txt")) if False else tarfile.open(fileobj=io.BytesIO(b"junk" * 200))
except tarfile.ReadError as e:
    print("ReadError", str(e).splitlines()[0])
try:
    bad = tarfile.TarInfo("../escape")
    bad.size = 0
    bb = io.BytesIO()
    with tarfile.open(fileobj=bb, mode="w") as tf:
        tf.addfile(bad)
    with tarfile.open(fileobj=io.BytesIO(bb.getvalue())) as tf:
        tf.extractall(os.path.join(d, "safe"), filter="data")
except tarfile.OutsideDestinationError as e:
    print("OutsideDestinationError", str(e).replace(d, "D"))
import shutil
shutil.rmtree(d)

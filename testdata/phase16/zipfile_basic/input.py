import io
import os
import tempfile
import zipfile

buf = io.BytesIO()
with zipfile.ZipFile(buf, "w", zipfile.ZIP_STORED) as z:
    zi = zipfile.ZipInfo("hello.txt", (2020, 5, 17, 10, 30, 12))
    zi.external_attr = 0o644 << 16
    z.writestr(zi, "hello zip\n")
    zi2 = zipfile.ZipInfo("dir/", (2020, 5, 17, 10, 30, 12))
    z.writestr(zi2, b"")
    zi3 = zipfile.ZipInfo("dir/ü.bin", (2021, 1, 2, 3, 4, 6))
    z.writestr(zi3, bytes(range(256)) * 4, compress_type=zipfile.ZIP_DEFLATED)
    z.comment = b"archive comment"
data = buf.getvalue()
print(len(data) > 100, data[:4], zipfile.is_zipfile(io.BytesIO(data)), zipfile.is_zipfile(io.BytesIO(b"nope")))
stored = io.BytesIO()
with zipfile.ZipFile(stored, "w") as z:
    zi = zipfile.ZipInfo("a.txt", (2000, 1, 1, 0, 0, 0))
    z.writestr(zi, "abc")
print(stored.getvalue().hex())

with zipfile.ZipFile(io.BytesIO(data)) as z:
    print(z.namelist(), z.comment)
    for info in z.infolist():
        print(info.filename, info.file_size, info.compress_type, info.date_time, oct(info.external_attr >> 16), info.is_dir(), hex(info.CRC), info.flag_bits)
    print(z.read("hello.txt"), z.read("dir/ü.bin")[:6], len(z.read("dir/ü.bin")), z.testzip())
    with z.open("hello.txt") as f:
        print(f.readline(), f.read(), f.name)
    print(z.getinfo("hello.txt"), repr(z.getinfo("dir/ü.bin")).startswith("<ZipInfo filename='dir/ü.bin' compress_type=deflate"))
    try:
        z.read("missing")
    except KeyError as e:
        print("KeyError", e)

d = tempfile.mkdtemp()
src = os.path.join(d, "src")
os.makedirs(os.path.join(src, "sub"))
with open(os.path.join(src, "one.txt"), "w") as f:
    f.write("one\n" * 100)
with open(os.path.join(src, "sub", "two.bin"), "wb") as f:
    f.write(bytes(range(200)))
path = os.path.join(d, "out.zip")
with zipfile.ZipFile(path, "w", zipfile.ZIP_DEFLATED) as z:
    z.write(os.path.join(src, "one.txt"), "one.txt")
    z.write(os.path.join(src, "sub"), "sub")
    z.write(os.path.join(src, "sub", "two.bin"), "sub/two.bin")
with zipfile.ZipFile(path, "a") as z:
    z.writestr("extra.txt", "added later")
with zipfile.ZipFile(path) as z:
    print(z.namelist(), z.getinfo("one.txt").compress_size < z.getinfo("one.txt").file_size)
    z.extractall(os.path.join(d, "x"))
    print(sorted(os.listdir(os.path.join(d, "x"))), open(os.path.join(d, "x", "sub", "two.bin"), "rb").read() == bytes(range(200)))
    print(z.extract("extra.txt", os.path.join(d, "y")).replace(d, "D"))
    print(z.open("one.txt").read().count(b"one"))
try:
    zipfile.ZipFile(io.BytesIO(b"junk" * 10))
except zipfile.BadZipFile as e:
    print("BadZipFile", e)
try:
    zipfile.ZipFile(os.path.join(d, "nope.zip"))
except FileNotFoundError as e:
    print("FileNotFoundError")
import shutil
shutil.rmtree(d)

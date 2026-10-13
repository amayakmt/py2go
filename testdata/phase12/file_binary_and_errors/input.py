import os
path = "/tmp/py2go_phase12_bin.dat"
with open(path, "wb") as f:
    print(f.write(b"\x00\x01\xfe\xffabc"), f.write(bytearray(b"!")))
with open(path, "rb") as f:
    data = f.read()
    print(data, len(data), type(data).__name__, f.read(), f.tell())
with open(path, "rb") as f:
    print(f.read(3), f.read(2), f.seek(1), f.read(1), f.seek(-2, 2), f.read())
with open(path, "r+b") as f:
    f.seek(2)
    f.write(b"XY")
    f.seek(0)
    print(f.read())
print(os.path.getsize(path), os.stat(path)[1] if False else "stat")
try:
    open("/tmp/py2go_does_not_exist/x.txt")
except FileNotFoundError as e:
    print(type(e).__name__, e)
try:
    open(path, "x")
except FileExistsError as e:
    print(type(e).__name__)
try:
    open("/tmp", "w")
except IsADirectoryError as e:
    print(type(e).__name__)
try:
    with open(path, "w") as f:
        f.write(b"bytes")
except TypeError as e:
    print(e)
try:
    with open(path, "wb") as f:
        f.write("text")
except TypeError as e:
    print(e)
try:
    open(path, "z")
except ValueError as e:
    print(e)
f = open(path, "rb")
f.close()
try:
    f.read()
except ValueError as e:
    print(e)
os.remove(path)
try:
    os.remove(path)
except FileNotFoundError as e:
    print(type(e).__name__, str(e).split(":")[0])

import os
import io

base = "/tmp/py2go_file_io_test"
os.makedirs(base, exist_ok=True)
p = base + "/t.txt"

with open(p, "w", encoding="utf-8", newline="") as f:
    print(f.write("héllo\r\nwörld\rlast\n"), f.tell(), repr(f), f.encoding, f.errors, f.mode, f.name == p, f.closed)
print(f.closed)
with open(p, "rb") as f:
    print(f.read(), type(f).__name__, repr(f).replace(base, "B"), f.mode, f.raw is f if hasattr(f, "raw") else "")
for nl in (None, "", "\n", "\r", "\r\n"):
    with open(p, newline=nl) as f:
        print(repr(nl), f.readlines(), f.newlines)
with open(p) as f:
    print(repr(f.read(3)), repr(f.read(2)), f.tell(), repr(f.readline()), repr(f.read()), repr(f.read()))
    f.seek(0)
    print(repr(f.readline(2)), repr(f.readline(100)), [l for l in f])
    f.seek(0, 2)
    print(f.tell(), f.seek(1), repr(f.read(1)))
with open(p, "w", newline="\r\n") as f:
    f.write("a\nb\n")
print(open(p, "rb").read())
with open(p, "a") as f:
    f.write("appended\n")
with open(p) as f:
    print(f.read().split("\n"))
with open(p, "r+") as f:
    f.write("XY")
    f.seek(0)
    print(repr(f.read(6)))
    f.truncate(3)
    f.seek(0)
    print(repr(f.read()))
with open(p, "w+") as f:
    f.write("round trip")
    f.seek(0)
    print(f.read(), f.readable(), f.writable(), f.seekable())
with open(p, "a+") as f:
    f.write("!!")
    f.seek(0)
    print(f.read())
try:
    open(p, "x")
except FileExistsError as e:
    print("x:", str(e).replace(base, "B"))
os.remove(p)
with open(p, "x") as f:
    f.write("fresh")

# encodings and errors
with open(p, "wb") as f:
    f.write(b"ok \xff\xfe bad")
for errs in ("replace", "ignore", "backslashreplace"):
    with open(p, encoding="utf-8", errors=errs) as f:
        print(errs, repr(f.read()))
try:
    open(p).read()
except UnicodeDecodeError as e:
    print(type(e).__name__, e)
with open(p, encoding="latin-1") as f:
    print(repr(f.read()))
with open(p, "w", encoding="ascii", errors="replace") as f:
    f.write("café")
print(open(p, "rb").read())
try:
    with open(p, "w", encoding="ascii") as f:
        f.write("café")
except UnicodeEncodeError as e:
    print(type(e).__name__, e)
with open(p, "w", encoding="utf-16") as f:
    f.write("hi")
print(open(p, "rb").read(), open(p, encoding="utf-16").read())
with open(p, "w", encoding="utf-8-sig") as f:
    f.write("bom")
print(open(p, "rb").read(), repr(open(p, encoding="utf-8-sig").read()), repr(open(p, encoding="utf-8").read()))

# binary
with open(p, "wb") as f:
    print(f.write(b"0123456789"), f.write(bytearray(b"ab")), type(f).__name__)
with open(p, "rb") as f:
    print(f.read(3), f.peek(2)[:2], f.read1(2), f.tell(), f.readline(), f.read(), f.read())
    b = bytearray(4)
    f.seek(2)
    print(f.readinto(b), b, f.seek(-2, 2), f.read())
with open(p, "r+b") as f:
    f.seek(5)
    f.write(b"XX")
    f.seek(0)
    print(f.read())
with open(p, "rb", buffering=0) as f:
    print(type(f).__name__, f.read(4), repr(f).replace(base, "B"))
with open(p, "rb") as f:
    print([bytes(l) for l in f], f.fileno() > 2, f.isatty())
fd = os.open(p, os.O_RDONLY)
with open(fd, "rb", closefd=True) as f:
    print(f.read(3), f.name == fd)
try:
    os.fstat(fd)
except OSError as e:
    print("fd closed", e.errno)

# errors
f = open(p)
f.close()
for op in (lambda: f.read(), lambda: f.write("x"), lambda: f.readline(), lambda: f.seek(0), lambda: f.flush(), lambda: next(f)):
    try:
        op()
    except ValueError as e:
        print(e)
f.close()
with open(p, "rb") as f:
    for op in (lambda: f.write(b"x"), lambda: f.read().decode() + "x"):
        try:
            op()
        except (io.UnsupportedOperation, OSError) as e:
            print(type(e).__name__, e)
with open(p) as f:
    try:
        f.write("x")
    except io.UnsupportedOperation as e:
        print(e)
for bad in [lambda: open(p, "rw"), lambda: open(p, "rbt"), lambda: open(p, "z"), lambda: open(p, "rb", encoding="utf-8"),
            lambda: open(p, "r", buffering=0), lambda: open(p, "rb", newline="\n"), lambda: open(p, "r", newline="x"),
            lambda: open(p, "rb", errors="strict")]:
    try:
        bad()
    except ValueError as e:
        print(e)
try:
    open(base)
except IsADirectoryError as e:
    print(str(e).replace(base, "B"))
try:
    open(base + "/nope/x", "w")
except FileNotFoundError as e:
    print(str(e).replace(base, "B"))

# text file wrappers
with open(p, "w") as f:
    print(f.buffer.write(b"via buffer"), f.line_buffering, f.buffer.name == p)
print(open(p).read())
import sys
print(sys.stdout.encoding, sys.stdout.errors, sys.stdout.name, sys.stdin.name, sys.stderr.name)
sys.stdout.write("direct write\n")
sys.stdout.buffer.write(b"binary stdout\n")
sys.stdout.flush()
os.remove(p)
os.rmdir(base)

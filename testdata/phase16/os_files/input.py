import os
import stat
import errno

base = "/tmp/py2go_os_files_test"
if os.path.exists(base):
    for root, dirs, files in os.walk(base, topdown=False):
        for f in files:
            os.remove(os.path.join(root, f))
        for d in dirs:
            os.rmdir(os.path.join(root, d))
    os.rmdir(base)

os.makedirs(os.path.join(base, "a", "b"))
os.makedirs(os.path.join(base, "a", "b"), exist_ok=True)
try:
    os.makedirs(os.path.join(base, "a", "b"))
except FileExistsError as e:
    print("exists:", e.errno == errno.EEXIST, e.strerror, e.filename == os.path.join(base, "a", "b"))
with open(os.path.join(base, "a", "one.txt"), "w") as f:
    f.write("hello\n")
with open(os.path.join(base, "a", "b", "two.bin"), "wb") as f:
    f.write(bytes(range(10)))
os.symlink("one.txt", os.path.join(base, "a", "link"))

print(sorted(os.listdir(os.path.join(base, "a"))))
for root, dirs, files in os.walk(base):
    dirs.sort()
    print(root[len(base):] or "/", sorted(dirs), sorted(files))

st = os.stat(os.path.join(base, "a", "one.txt"))
print(st.st_size, stat.S_ISREG(st.st_mode), stat.S_ISDIR(st.st_mode), oct(stat.S_IMODE(st.st_mode)), st.st_nlink, len(st), st[6])
lst = os.lstat(os.path.join(base, "a", "link"))
print(stat.S_ISLNK(lst.st_mode), os.readlink(os.path.join(base, "a", "link")), os.path.islink(os.path.join(base, "a", "link")))
print(stat.filemode(st.st_mode)[0], stat.filemode(os.stat(base).st_mode)[0])

entries = sorted(os.scandir(os.path.join(base, "a")), key=lambda e: e.name)
print([(e.name, e.is_dir(), e.is_file(), e.is_symlink()) for e in entries], repr(entries[0]))
print(entries[0].path == os.path.join(base, "a", "b"), entries[1].stat().st_size)

os.rename(os.path.join(base, "a", "one.txt"), os.path.join(base, "a", "uno.txt"))
os.chmod(os.path.join(base, "a", "uno.txt"), 0o600)
print(oct(stat.S_IMODE(os.stat(os.path.join(base, "a", "uno.txt")).st_mode)))
os.utime(os.path.join(base, "a", "uno.txt"), (1000000000, 1500000000))
print(os.stat(os.path.join(base, "a", "uno.txt")).st_mtime, os.stat(os.path.join(base, "a", "uno.txt")).st_atime_ns)
os.truncate(os.path.join(base, "a", "uno.txt"), 2)
print(os.path.getsize(os.path.join(base, "a", "uno.txt")))

fd = os.open(os.path.join(base, "fd.txt"), os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o644)
print(os.write(fd, b"abcdef"), os.lseek(fd, 0, os.SEEK_CUR), os.fstat(fd).st_size)
os.close(fd)
fd = os.open(os.path.join(base, "fd.txt"), os.O_RDONLY)
print(os.read(fd, 3), os.read(fd, 100), os.read(fd, 10), os.isatty(fd))
os.close(fd)
r, w = os.pipe()
os.write(w, b"ping")
os.close(w)
print(os.read(r, 10))
os.close(r)

for bad in [lambda: os.stat(base + "/nope"), lambda: os.listdir(base + "/nope"), lambda: os.mkdir(base),
            lambda: os.rmdir(base), lambda: os.remove(base), lambda: os.rename(base + "/x", base + "/y"),
            lambda: os.close(9999), lambda: os.open(base + "/a/uno.txt", os.O_WRONLY | os.O_CREAT | os.O_EXCL),
            lambda: os.listdir(base + "/a/uno.txt"), lambda: os.readlink(base + "/a/uno.txt")]:
    try:
        bad()
    except OSError as e:
        print(type(e).__name__, e.errno, str(e).replace(base, "BASE"))

print(os.path.exists(base), os.path.isdir(base), os.path.isfile(base), os.access(base, os.R_OK), os.access(base + "/nope", os.F_OK))
print(os.path.realpath(base + "/a/link").replace(base, "BASE"), os.path.relpath(base + "/a/b", base), os.path.samefile(base, base + "/a/.."))
print(os.fspath(base) == base, os.fsencode("é"), os.fsdecode(b"x"), os.name, os.sep, os.linesep == "\n", os.curdir, os.devnull)
os.environ["PY2GO_X"] = "1"
print(os.environ["PY2GO_X"], os.getenv("PY2GO_X"), os.getenv("PY2GO_NOPE", "dflt"), "PY2GO_X" in os.environ)
del os.environ["PY2GO_X"]
print(os.getenv("PY2GO_X"), os.environ.get("PY2GO_X"))
try:
    os.environ["a"] = 1
except TypeError as e:
    print(e)
old = os.getcwd()
os.chdir(base)
print(os.getcwd() == os.path.realpath(base), os.path.abspath("a") == base + "/a")
os.chdir(old)
print(os.strerror(2), os.strerror(13), os.cpu_count() >= 1, isinstance(os.getpid(), int), os.get_terminal_size.__name__)
print(os.uname().sysname, len(os.uname()), os.umask(0o22) is not None)

for root, dirs, files in os.walk(base, topdown=False):
    for f in files:
        os.remove(os.path.join(root, f))
    for d in dirs:
        p = os.path.join(root, d)
        if os.path.islink(p):
            os.remove(p)
        else:
            os.rmdir(p)
os.rmdir(base)
print(os.path.exists(base))

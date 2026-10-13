import os
from os import path
import os.path as osp
print(path.join("a", "b"), path.join("a", "/b", "c"), path.join("a/", "b"), path.join("", "x"), path.join("a", ""), osp.join("/", "usr", "lib"))
for p in ("/usr/lib/python.py", "file.tar.gz", "/", "", "a/b/", ".hidden", "dir/.hidden.txt", "noext", "a.b/c"):
    print(repr(p), path.basename(p), path.dirname(p), path.split(p), path.splitext(p), path.isabs(p))
for p in ("a/./b/../c", "/../a", "//a///b", "", ".", "..", "a/b/../../..", "/a/b/./c/.."):
    print(repr(p), path.normpath(p))
print(path.commonprefix(["/usr/lib", "/usr/local"]), path.commonpath(["/usr/lib", "/usr/local"]), path.relpath("/a/b/c", "/a"), path.relpath("/a", "/a/b/c"), path.relpath("/a/b", "/a/b"))
print(os.sep, os.pathsep, os.linesep == "\n", os.curdir, os.pardir, os.name, path.sep, os.extsep)
print(path.exists("/tmp"), path.isdir("/tmp"), path.isfile("/tmp"), path.exists("/definitely/not/here"), path.isabs(os.getcwd()), path.abspath("/a/../b"), path.expanduser("/x"))
os.makedirs("/tmp/py2go_p12/a/b", exist_ok=True)
os.makedirs("/tmp/py2go_p12/a/b", exist_ok=True)
for n in ("one.txt", "two.txt"):
    with open(path.join("/tmp/py2go_p12/a", n), "w") as f:
        f.write(n)
print(sorted(os.listdir("/tmp/py2go_p12/a")), [(r.replace("/tmp/py2go_p12", "."), sorted(d), sorted(fs)) for r, d, fs in os.walk("/tmp/py2go_p12")])
os.rename("/tmp/py2go_p12/a/one.txt", "/tmp/py2go_p12/a/uno.txt")
print(sorted(os.listdir("/tmp/py2go_p12/a")))
try:
    os.makedirs("/tmp/py2go_p12/a/b")
except FileExistsError as e:
    print(type(e).__name__)
try:
    os.mkdir("/tmp/py2go_p12/a")
except FileExistsError as e:
    print("exists")
try:
    os.listdir("/tmp/py2go_p12/missing")
except FileNotFoundError as e:
    print(type(e).__name__)
for n in ("uno.txt", "two.txt"):
    os.remove(path.join("/tmp/py2go_p12/a", n))
os.rmdir("/tmp/py2go_p12/a/b"); os.rmdir("/tmp/py2go_p12/a"); os.rmdir("/tmp/py2go_p12")
print(path.exists("/tmp/py2go_p12"), os.getenv("PY2GO_SURELY_UNSET"), os.getenv("PY2GO_SURELY_UNSET", "dflt"), "PATH" in os.environ, type(os.getpid()).__name__, os.cpu_count() >= 1)

import os
import shutil
import tempfile

d = tempfile.mkdtemp(prefix="py2go_")
print(os.path.basename(d).startswith("py2go_"), os.path.isdir(d), oct(os.stat(d).st_mode & 0o777))
src = os.path.join(d, "src")
os.makedirs(os.path.join(src, "inner"))
with open(os.path.join(src, "a.txt"), "w") as f:
    f.write("alpha")
with open(os.path.join(src, "inner", "b.log"), "w") as f:
    f.write("beta")
os.symlink("a.txt", os.path.join(src, "ln"))

print(shutil.copy(os.path.join(src, "a.txt"), os.path.join(d, "copy.txt")).replace(d, "D"))
print(shutil.copyfile(os.path.join(src, "a.txt"), os.path.join(d, "cf.txt")).replace(d, "D"), open(os.path.join(d, "cf.txt")).read())
os.chmod(os.path.join(src, "a.txt"), 0o640)
shutil.copy2(os.path.join(src, "a.txt"), os.path.join(d, "c2.txt"))
print(oct(os.stat(os.path.join(d, "c2.txt")).st_mode & 0o777), os.stat(os.path.join(d, "c2.txt")).st_mtime == os.stat(os.path.join(src, "a.txt")).st_mtime)
try:
    shutil.copyfile(os.path.join(src, "a.txt"), os.path.join(src, "a.txt"))
except shutil.SameFileError as e:
    print("same:", str(e).replace(d, "D"))
shutil.copytree(src, os.path.join(d, "dst"), symlinks=True)
print(sorted(os.listdir(os.path.join(d, "dst"))), os.path.islink(os.path.join(d, "dst", "ln")))
shutil.copytree(src, os.path.join(d, "dst2"), ignore=shutil.ignore_patterns("*.log", "ln"))
print(sorted(os.listdir(os.path.join(d, "dst2"))), sorted(os.listdir(os.path.join(d, "dst2", "inner"))))
try:
    shutil.copytree(src, os.path.join(d, "dst"))
except FileExistsError as e:
    print("exists:", type(e).__name__)
shutil.copytree(src, os.path.join(d, "dst"), dirs_exist_ok=True)
print(shutil.move(os.path.join(d, "copy.txt"), os.path.join(d, "moved.txt")).replace(d, "D"), os.path.exists(os.path.join(d, "copy.txt")))
print(shutil.move(os.path.join(d, "moved.txt"), os.path.join(d, "dst2")).replace(d, "D"))
shutil.rmtree(os.path.join(d, "dst"))
print(os.path.exists(os.path.join(d, "dst")))
errors = []
shutil.rmtree(os.path.join(d, "missing"), onerror=lambda func, path, exc: errors.append((func.__name__, path.replace(d, "D"), exc[0].__name__)))
print(errors)
shutil.rmtree(os.path.join(d, "missing"), ignore_errors=True)
try:
    shutil.rmtree(os.path.join(d, "missing"))
except FileNotFoundError as e:
    print("rmtree:", str(e).replace(d, "D"))
print(shutil.which("sh") is not None, shutil.which("definitely-not-here"), shutil.disk_usage("/").total > 0, shutil.get_terminal_size((80, 24)).columns > 0)

fd, path = tempfile.mkstemp(suffix=".txt", prefix="pre_", dir=d)
os.write(fd, b"mkstemp")
os.close(fd)
print(os.path.basename(path)[:4], path.endswith(".txt"), open(path).read(), os.path.dirname(path) == d, oct(os.stat(path).st_mode & 0o777))
with tempfile.NamedTemporaryFile(dir=d) as tf:
    tf.write(b"named")
    tf.flush()
    print(os.path.exists(tf.name), os.path.dirname(tf.name) == d, open(tf.name, "rb").read(), tf.mode, tf.name == tf.file.name)
    tf.seek(0)
    print(tf.read())
print(os.path.exists(tf.name))
with tempfile.NamedTemporaryFile("w+", suffix=".csv", delete=False, dir=d) as tf:
    tf.write("a,b\n")
    name = tf.name
print(os.path.exists(name), open(name).read(), name.endswith(".csv"))
with tempfile.TemporaryFile("w+") as tf:
    tf.write("anon")
    tf.seek(0)
    print(tf.read())
with tempfile.TemporaryDirectory(dir=d) as td:
    print(os.path.isdir(td), os.path.dirname(td) == d)
    open(os.path.join(td, "x"), "w").close()
print(os.path.exists(td))
tdo = tempfile.TemporaryDirectory(prefix="obj_", dir=d)
print(os.path.exists(tdo.name), repr(tdo).startswith("<TemporaryDirectory '"))
tdo.cleanup()
print(os.path.exists(tdo.name))
with tempfile.SpooledTemporaryFile(max_size=10, mode="w+") as sp:
    sp.write("short")
    print(sp._rolled, sp.tell())
    sp.write(" and a much longer text")
    print(sp._rolled)
    sp.seek(0)
    print(sp.read())
print(tempfile.gettempdir() != "", tempfile.gettempprefix(), os.path.isabs(tempfile.mktemp()), tempfile.mktemp(suffix="s", prefix="p", dir="/x").startswith("/x/p"))
shutil.rmtree(d)
print(os.path.exists(d))

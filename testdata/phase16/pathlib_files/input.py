from pathlib import Path
import os
import shutil

base = Path("/tmp/py2go_pathlib_test")
if base.exists():
    shutil.rmtree(base)
base.mkdir()
(base / "sub" / "deep").mkdir(parents=True)
try:
    (base / "sub").mkdir()
except FileExistsError as e:
    print("exists:", str(e).replace(str(base), "BASE"))
(base / "sub").mkdir(exist_ok=True)
try:
    (base / "x" / "y").mkdir()
except FileNotFoundError as e:
    print("nf:", str(e).replace(str(base), "BASE"))
f = base / "hello.txt"
print(f.write_text("héllo\nworld\n"), f.read_text(), f.read_bytes(), f.stat().st_size, f.exists(), f.is_file(), f.is_dir())
print(f.write_bytes(b"\x00\x01"), f.read_bytes())
with f.open("w") as fh:
    fh.write("opened")
print(f.read_text(), f.name, f.suffix)
(base / "a.py").touch()
(base / "sub" / "b.py").touch()
(base / "sub" / "deep" / "c.py").write_text("c")
(base / "sub" / "d.txt").write_text("d")
print(sorted(p.name for p in base.iterdir()))
print(sorted(str(p.relative_to(base)) for p in base.glob("*.py")))
print(sorted(str(p.relative_to(base)) for p in base.glob("*")))
print(sorted(str(p.relative_to(base)) for p in base.rglob("*.py")))
print(sorted(str(p.relative_to(base)) for p in base.glob("**/*.py")))
print(sorted(str(p.relative_to(base)) for p in base.glob("sub/*")))
print(sorted(str(p.relative_to(base)) for p in base.glob("**")))
print(sorted(str(p.relative_to(base)) for p in base.glob("*/")))
print(sorted(str(p.relative_to(base)) for p in base.rglob("d.*")), sorted(p.name for p in (base / "sub").glob("[bd]*")))
print(sorted(str(p.relative_to(base)) for p in base.glob("nomatch*")), list(base.glob("hello.txt")) == [f])
for root, dirs, files in base.walk():
    dirs.sort()
    print(str(root.relative_to(base)), dirs, sorted(files))
src = base / "a.py"
dst = src.rename(base / "z.py")
print(dst.name, src.exists(), dst.exists())
dst2 = dst.replace(base / "y.py")
print(dst2.name, sorted(p.name for p in base.glob("*.py")))
(base / "link").symlink_to("hello.txt")
print((base / "link").is_symlink(), (base / "link").readlink(), (base / "link").resolve() == f.resolve(), (base / "link").read_text())
print(f.samefile(base / "link"), f.chmod(0o600), oct(f.stat().st_mode & 0o777), f.lstat().st_size, (base / "link").lstat().st_size)
(base / "y.py").unlink()
try:
    (base / "y.py").unlink()
except FileNotFoundError as e:
    print("unlink:", str(e).replace(str(base), "BASE"))
(base / "y.py").unlink(missing_ok=True)
try:
    (base / "sub").rmdir()
except OSError as e:
    print("rmdir:", e.errno)
(base / "sub" / "deep" / "c.py").unlink()
(base / "sub" / "deep").rmdir()
try:
    (base / "touch.txt").touch(exist_ok=False)
    (base / "touch.txt").touch(exist_ok=False)
except FileExistsError as e:
    print("touch:", type(e).__name__)
print(f.owner() == Path(f).owner(), isinstance(f.group(), str))
shutil.rmtree(base)
print(base.exists())

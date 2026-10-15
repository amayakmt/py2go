import os
import shutil
import tempfile
import zipfile
import tarfile

d = tempfile.mkdtemp()
src = os.path.join(d, "proj")
os.makedirs(os.path.join(src, "pkg"))
for rel, text in (("a.txt", "alpha"), ("pkg/b.txt", "beta")):
    with open(os.path.join(src, rel), "w") as f:
        f.write(text)
print([n for n, _ in shutil.get_archive_formats() if n in ('tar', 'gztar', 'zip')], [n for n, _, _ in shutil.get_unpack_formats() if n in ('tar', 'gztar', 'zip')])
z = shutil.make_archive(os.path.join(d, "out_zip"), "zip", root_dir=d, base_dir="proj")
t = shutil.make_archive(os.path.join(d, "out_tgz"), "gztar", root_dir=d, base_dir="proj")
tt = shutil.make_archive(os.path.join(d, "out_tar"), "tar", root_dir=src)
print([os.path.basename(x) for x in (z, t, tt)])
print(sorted(zipfile.ZipFile(z).namelist()), sorted(tarfile.open(t).getnames()), sorted(tarfile.open(tt).getnames()))
for archive in (z, t, tt):
    out = os.path.join(d, "unp_" + os.path.basename(archive))
    shutil.unpack_archive(archive, out)
    found = sorted(os.path.relpath(os.path.join(r, f), out) for r, _, fs in os.walk(out) for f in fs)
    print(found)
try:
    shutil.make_archive(os.path.join(d, "x"), "rar", root_dir=d)
except ValueError as e:
    print(e)
try:
    shutil.unpack_archive(os.path.join(d, "a.unknown"))
except shutil.ReadError as e:
    print(str(e).replace(d, "D"))
shutil.rmtree(d)

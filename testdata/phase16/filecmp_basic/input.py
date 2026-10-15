import filecmp
import os
import tempfile
import shutil

d = "/tmp/py2go_filecmp_test"
shutil.rmtree(d, ignore_errors=True)
os.makedirs(d)
a = os.path.join(d, "a")
b = os.path.join(d, "b")
for root in (a, b):
    os.makedirs(os.path.join(root, "sub"))
files = {"same.txt": ("same", "same"), "diff.txt": ("left", "right!"), "left_only.txt": ("l", None), "right_only.txt": (None, "r"),
         "sub/inner.txt": ("in", "in")}
for name, (l, r) in files.items():
    if l is not None:
        open(os.path.join(a, name), "w").write(l)
    if r is not None:
        open(os.path.join(b, name), "w").write(r)
print(filecmp.cmp(os.path.join(a, "same.txt"), os.path.join(b, "same.txt")), filecmp.cmp(os.path.join(a, "diff.txt"), os.path.join(b, "diff.txt")),
      filecmp.cmp(os.path.join(a, "same.txt"), os.path.join(b, "diff.txt"), shallow=False))
print(filecmp.cmpfiles(a, b, ["same.txt", "diff.txt", "left_only.txt", "sub"]))
dc = filecmp.dircmp(a, b)
print(sorted(dc.left_only), sorted(dc.right_only), sorted(dc.common), sorted(dc.same_files), sorted(dc.diff_files), sorted(dc.common_dirs), sorted(dc.common_files))
print(sorted(dc.subdirs), sorted(dc.subdirs["sub"].same_files))
dc.report()
shutil.rmtree(d)

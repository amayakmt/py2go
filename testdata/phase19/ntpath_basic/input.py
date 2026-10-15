import ntpath
import nturl2path

print(ntpath.join("C:\\a", "b", "c.txt"), ntpath.join("C:", "x"), ntpath.join("a", "\\b"), ntpath.join("C:\\a", "D:\\b"))
print(ntpath.join(b"C:\\a", b"b"), ntpath.join("//host/share", "x"))
print(ntpath.split("C:\\a\\b\\c.txt"), ntpath.split("C:\\"), ntpath.split("\\\\h\\s\\x\\y"))
print(ntpath.splitdrive("C:\\a\\b"), ntpath.splitdrive("//h/s/a"), ntpath.splitdrive("a\\b"))
print(ntpath.splitroot("C:\\a\\b"), ntpath.splitroot("C:a"), ntpath.splitroot("\\a"), ntpath.splitroot("//h/s/a"))
print(ntpath.splitext("a\\b.tar.gz"), ntpath.splitext("a.b\\c"), ntpath.splitext(".bashrc"), ntpath.splitext(b"x.py"))
print(ntpath.basename("C:\\a\\b.txt"), ntpath.dirname("C:\\a\\b.txt"), ntpath.basename("a/b/"))
print(ntpath.normpath("C:/a//b/./c/../d"), ntpath.normpath("..\\..\\a"), ntpath.normpath(""), ntpath.normpath("C:\\.."), ntpath.normpath(b"a//b"))
print(ntpath.normcase("C:/Foo/BAR"), ntpath.isabs("C:\\a"), ntpath.isabs("C:a"), ntpath.isabs("\\\\h\\s"), ntpath.isabs("a"))
print(ntpath.commonpath(["C:\\a\\b", "C:\\a\\c"]), ntpath.commonprefix(["C:\\ab", "C:\\ac"]))
print(ntpath.relpath("C:\\a\\b\\c", "C:\\a"), ntpath.relpath("C:\\a", "C:\\a\\b\\c"), ntpath.relpath("C:\\a", "C:\\a"))
print(ntpath.isreserved("CON"), ntpath.isreserved("a\\nul.txt"), ntpath.isreserved("ok"))
print(ntpath.sep, ntpath.altsep, ntpath.pathsep, ntpath.devnull, ntpath.curdir)
print(ntpath.expandvars("$$ and %NOPE_X% and ${NOPE_Y}"))
print(nturl2path.url2pathname("///C:/Users/x/a%20b.txt"), nturl2path.pathname2url("C:\\Users\\x\\a b.txt"))
print(nturl2path.pathname2url("\\\\h\\s\\f"), nturl2path.url2pathname("/a/b"))
try:
    ntpath.join("a", b"b")
except TypeError as e:
    print("TypeError", e)
try:
    ntpath.commonpath(["C:\\a", "D:\\a"])
except ValueError as e:
    print("ValueError", e)

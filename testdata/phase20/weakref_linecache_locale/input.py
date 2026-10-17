import linecache
import locale
import os
import tempfile
import weakref


class Node:
    def __init__(self, name):
        self.name = name


n = Node("a")
r = weakref.ref(n)
print(r() is n, r().name, r() == weakref.ref(n))
calls = []
f = weakref.finalize(n, calls.append, "done")
print(f.alive, f.peek() is not None)
f()
print(f.alive, calls, f())

wvd = weakref.WeakValueDictionary()
wvd["k"] = n
print(len(wvd), "k" in wvd, wvd.get("k").name, list(wvd))
wkd = weakref.WeakKeyDictionary()
wkd[n] = 5
print(wkd[n], len(wkd))
ws = weakref.WeakSet([n])
ws.add(n)
print(len(ws), n in ws)
ws.discard(n)
print(len(ws))

d = tempfile.mkdtemp()
p = os.path.join(d, "src.py")
with open(p, "w") as fh:
    fh.write("line one\nline two\nlast")
print(repr(linecache.getline(p, 1)), repr(linecache.getline(p, 3)), repr(linecache.getline(p, 9)))
print(len(linecache.getlines(p)))
linecache.clearcache()
print(repr(linecache.getline(os.path.join(d, "missing.py"), 1)))
os.remove(p)
os.rmdir(d)

print(locale.setlocale(locale.LC_ALL, "C"), locale.getlocale(), locale.localeconv()["decimal_point"])
print(locale.atof("3.5"), locale.atoi("7"), locale.format_string("%d items", 3), locale.strcoll("a", "b"))

import mimetypes
import netrc
import os
import tempfile

print(mimetypes.guess_type("a/b/pic.png"), mimetypes.guess_type("x.tar.gz"), mimetypes.guess_type("page.html"), mimetypes.guess_type("noext"))
print(mimetypes.guess_type("http://h.com/d/f.json?x=1#y"), mimetypes.guess_type("data:text/plain;base64,aGk="), mimetypes.guess_type("a.JPG"))
print(mimetypes.guess_extension("text/html"), mimetypes.guess_extension("image/jpeg"), sorted(mimetypes.guess_all_extensions("text/plain"))[:3])
print(mimetypes.guess_type("x.unknownext"), mimetypes.guess_type("a.mp3", strict=False))
mimetypes.add_type("application/x-foo", ".foo")
print(mimetypes.guess_type("t.foo"), mimetypes.guess_extension("application/x-foo"))

d = tempfile.mkdtemp()
p = os.path.join(d, "netrc")
with open(p, "w") as f:
    f.write("machine example.com login alice password s3cret\ndefault login anon password guest\n")
os.chmod(p, 0o600)
n = netrc.netrc(p)
print(n.authenticators("example.com"), n.authenticators("other.org"), sorted(n.hosts))
print(repr(n).splitlines()[0])
with open(p, "w") as f:
    f.write("machine x login\n")
try:
    netrc.netrc(p)
except netrc.NetrcParseError as e:
    print(str(e).replace(d, "D"))
os.remove(p)
os.rmdir(d)

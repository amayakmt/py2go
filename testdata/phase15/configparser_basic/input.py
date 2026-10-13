import configparser
import io

TEXT = """
[DEFAULT]
base = /opt
debug = no

[paths]
home = %(base)s/home
logs: %(home)s/logs
multi = line one
    line two

    line four
empty =
; a comment
# another
inline = value ; not a comment

[server]
host = example.com
port = 8080
ratio = 0.75
enabled = yes
Name With Spaces = x
"""
cp = configparser.ConfigParser()
cp.read_string(TEXT)
print(cp.sections(), cp.defaults(), cp.options("paths"), cp.has_section("server"), cp.has_option("server", "host"), cp.has_option("server", "base"))
print(cp.get("paths", "home"), cp.get("paths", "logs"), repr(cp.get("paths", "multi")), repr(cp.get("paths", "empty")), cp.get("paths", "inline"), cp.get("paths", "logs", raw=True))
print(cp.getint("server", "port"), cp.getfloat("server", "ratio"), cp.getboolean("server", "enabled"), cp.getboolean("DEFAULT", "debug"), cp.get("server", "name with spaces"), cp.get("server", "missing", fallback="dflt"), cp.getint("server", "nope", fallback=7), cp.get("nosuch", "x", fallback=None))
print(cp["server"]["host"], cp["server"].get("port"), cp["server"].getint("port"), cp["server"].get("zzz", "fb"), "host" in cp["server"], "zzz" in cp["server"], list(cp["server"]), len(cp["server"]), cp["paths"].name, cp["server"], dict(cp["server"]), "server" in cp, "nope" in cp, len(cp), list(cp))
print(cp.items("server"), cp.items("paths", raw=True)[:2], [s for s, _ in cp.items()])
cp.set("server", "port", "9090")
cp.set("server", "new", "val")
cp["extra"] = {"a": "1", "b": "2"}
cp["server"]["added"] = "yes"
cp.add_section("blank")
cp.remove_option("server", "ratio")
cp.remove_section("blank")
del cp["server"]["added"]
out = io.StringIO()
cp.write(out)
print(out.getvalue())
out2 = io.StringIO()
cp.write(out2, space_around_delimiters=False)
print(out2.getvalue()[:60])
for call in (lambda: cp.get("nosuch", "x"), lambda: cp.get("server", "nosuch"), lambda: cp.getint("server", "host"), lambda: cp.getboolean("server", "host"), lambda: cp.add_section("server"), lambda: cp.add_section("DEFAULT"), lambda: cp.set("server", "x", 5), lambda: cp.set("server", "x", "%bad"), lambda: cp["nosuch"], lambda: cp["server"]["nosuch"], lambda: cp.options("nosuch"), lambda: cp.remove_option("nosuch", "x"), lambda: cp.set("nosuch", "a", "b")):
    try:
        print(call())
    except (configparser.Error, ValueError, TypeError, KeyError) as e:
        print(type(e).__name__, e)
bad = [("no header\n", {}), ("[a]\nx=1\n[a]\ny=2\n", {}), ("[a]\nx=1\nx=2\n", {}), ("[a]\nx=1\nbadline\n", {}), ("[a]\n=nokey\n", {}), ("[a]\nk=%(missing)s\n", {}), ("[a]\nk=%(k)s\n", {}), ("[a]\nk=%x\n", {})]
for text, kw in bad:
    p = configparser.ConfigParser(**kw)
    try:
        p.read_string(text)
        print({s: dict(p[s]) for s in p.sections()})
    except (configparser.Error, ValueError) as e:
        print(type(e).__name__, str(e).replace("\n", "|"))
p = configparser.ConfigParser(strict=False)
p.read_string("[a]\nx=1\nx=2\n[a]\ny=3\n")
print(dict(p["a"]))
ext = configparser.ConfigParser(interpolation=configparser.ExtendedInterpolation())
ext.read_string("[a]\nx = 1\n[b]\ny = ${a:x}-${z}\nz = last\nw = $$literal\n")
print(ext["b"]["y"], ext["b"]["w"])
try:
    ext.read_string("[c]\nq = ${nosuch:x}\n")
    print(ext["c"]["q"])
except configparser.Error as e:
    print(type(e).__name__, e)
nv = configparser.ConfigParser(allow_no_value=True, delimiters=("=",), comment_prefixes=("#",), inline_comment_prefixes=(";",))
nv.read_string("[s]\nflag\nkey = v ; trailing\n# comment\nother: not-a-delim\n")
print(dict(nv["s"]), nv.get("s", "flag"))
raw = configparser.RawConfigParser(defaults={"k": "v%"}, default_section="COMMON")
raw.read_string("[x]\na = %(k)s\n")
print(raw.get("x", "a"), raw.defaults(), raw.get("x", "k"))
rd = configparser.ConfigParser(allow_no_value=True)
rd.read_dict({"s1": {"a": 1, "b": None}, "s2": {"c": True}})
print({s: dict(rd[s]) for s in rd.sections()})
print(rd.read(["/nonexistent/file.ini"]))
with open("/tmp/py2go_cfg_test.ini", "w") as f:
    f.write("[one]\nk = v\n")
fp = configparser.ConfigParser()
print(fp.read("/tmp/py2go_cfg_test.ini"), fp.sections(), fp["one"]["k"])
with open("/tmp/py2go_cfg_test.ini") as f:
    fp2 = configparser.ConfigParser()
    fp2.read_file(f)
    print(fp2.sections())
print(repr(configparser.NoSectionError("s")), configparser.DEFAULTSECT, issubclass(configparser.NoSectionError, configparser.Error), issubclass(configparser.Error, Exception))

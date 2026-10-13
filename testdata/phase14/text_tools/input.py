import fnmatch
import getopt
import glob
import os
import secrets
import shlex
import colorsys

print(fnmatch.fnmatch("main.py", "*.py"), fnmatch.fnmatch("main.PY", "*.py"), fnmatch.fnmatchcase("a.txt", "[a-c].txt"), fnmatch.fnmatch("d.txt", "[!a-c].txt"), fnmatch.fnmatch("x", "?"), fnmatch.fnmatch("a/b", "a*b"))
print(fnmatch.filter(["a.py", "b.txt", "c.py", ".hidden.py"], "*.py"), fnmatch.translate("*.py"), fnmatch.translate("[a-c]?x"))
print(shlex.split("a 'b c' \"d e\" f\\ g # comment"), shlex.split("a b # c", comments=True), shlex.split("x=1 y='2 3'"), shlex.quote("it's"), shlex.quote("safe-1.2"), shlex.quote(""), shlex.join(["ls", "a b", "c"]))
try:
    shlex.split("'unclosed")
except ValueError as e:
    print("ValueError", e)
lex = shlex.shlex("foo = 'bar baz' ; qux", posix=True)
print(list(lex))
args = ["-a", "-b", "val", "--long=1", "--flag", "rest", "-x"]
print(getopt.getopt(args, "ab:", ["long=", "flag"]), getopt.gnu_getopt(args, "ab:x", ["long=", "flag"]))
def attempt(argv):
    try:
        getopt.getopt(argv, "ab:", ["flag"])
    except getopt.GetoptError as e:
        print(type(e).__name__, e, e.msg, e.opt)


attempt(["-z"])
attempt(["-b"])
attempt(["--nope"])
attempt(["--flag=1"])
print(colorsys.rgb_to_hsv(0.2, 0.4, 0.4), colorsys.hsv_to_rgb(0.5, 0.5, 0.4), colorsys.rgb_to_hls(0.2, 0.4, 0.4), colorsys.hls_to_rgb(0.5, 0.3, 0.3), colorsys.rgb_to_yiq(0.1, 0.5, 0.9), colorsys.yiq_to_rgb(0.5, 0.1, -0.1))
print(len(secrets.token_bytes(7)), len(secrets.token_hex(5)), len(secrets.token_urlsafe(9)), 0 <= secrets.randbelow(10) < 10, secrets.choice("abc") in "abc", secrets.compare_digest("a", "a"), secrets.compare_digest(b"ab", b"ac"))
base = "/tmp/py2go_glob_test"
os.makedirs(base + "/sub/deep", exist_ok=True)
for name in ["a.py", "b.py", "c.txt", ".h.py", "sub/d.py", "sub/e.txt", "sub/deep/f.py"]:
    with open(os.path.join(base, name), "w") as f:
        f.write("x")
rel = lambda paths: sorted(p[len(base) + 1:] for p in paths)
print(rel(glob.glob(base + "/*.py")), rel(glob.glob(base + "/*")), rel(glob.glob(base + "/**/*.py", recursive=True)), rel(glob.glob(base + "/*/*.txt")), rel(glob.glob(base + "/.*.py")))
print(rel(glob.glob(base + "/[ab].py")), rel(glob.glob(base + "/nomatch*")), glob.has_magic("a*"), glob.has_magic("abc"), glob.escape("a*b[c]"), sorted(glob.glob("*.py", root_dir=base)))

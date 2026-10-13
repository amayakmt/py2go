import re
for p in ("(", ")", "(a", "a)", "[a", "a**", "a+*", "*a", "+", "?", "{1}a", "a{2,1}", "\\", "\\q", "(?P<n>a)(?P<n>b)", "(?P=zz)", "\\1", "(?<=a*)b", "[b-a]", "(?z)", "(?P<1a>x)", "(?P<>x)", "a{1", "(?i", "(?#c", "\\x4", "\\u12", "(?(1)a)", "(?P<n", "[\\d-z]", "x{1,2}{3}", "a|*"):
    try:
        re.compile(p)
        print(repr(p), "ok")
    except re.error as e:
        print(repr(p), e.msg, e.pos, e)
print(re.escape("a.b*c?d+e(f)g[h]i{j}k|l^m$n\\o p-q#r&s~t"), re.escape("abc_123"), re.escape(""), re.escape("é日"), re.escape(b"a.b"), re.escape("a\nb\t"))
print(re.compile("a{").match("a{") is not None, re.compile("{").match("{") is not None, re.compile("a{x}").match("a{x}") is not None, re.compile("a{,}").match("a{,}") is not None, re.compile("a{1,2").match("a{1,2") is not None)
try:
    re.match("a", 5)
except TypeError as e:
    print(e)
try:
    re.match(5, "a")
except TypeError as e:
    print(e)
try:
    re.match(b"a", "a")
except TypeError as e:
    print(e)
try:
    re.match("a", b"a")
except TypeError as e:
    print(e)
try:
    re.compile(re.compile("a"), re.I)
except ValueError as e:
    print(e)
print(re.compile(re.compile("a")).pattern, issubclass(re.error, Exception), re.error.__name__ if hasattr(re.error, "__name__") else "error")

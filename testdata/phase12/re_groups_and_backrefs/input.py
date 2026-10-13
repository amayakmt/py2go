import re
m = re.match(r"(a)(b)?(c)", "ac")
print(m.groups(), m.groups("-"), m.group(2), m.group(1, 3), m.span(2), m.start(2), m.end(2), m.lastindex, m.regs)
m = re.match(r"(?P<first>\w+) (?P<last>\w+)", "Jane Doe!")
print(m.group("first"), m["last"], m.groupdict(), m.span("last"), m.lastgroup, m.group(0), m.expand(r"\g<last>, \g<first> (\1)"), m.pos, m.endpos, m.re.pattern)
print(re.match(r"(a)\1", "aa").group(), re.match(r"(a)\1", "ab"), re.match(r"(?P<x>a)(?P=x)", "aa").group(), re.match(r"(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)\10", "abcdefghijj").group(), re.match(r"(?i)(a)\1", "aA").group())
print(re.findall(r"(a)|(b)", "ab"), re.findall(r"(a)(b)?", "aab"), re.search(r"(a)*", "aaa").group(1), re.search(r"(a|b)*", "abab").groups(), re.search(r"((a)|(b))+", "ab").groups())
print(re.match(r"(?:(a)|b)+", "ab").groups(), re.match(r"(a)?b", "b").groups(), re.match(r"(a)?b", "b").group(1), re.match(r"(a)?b", "b").span(1), re.match(r"((a)(b))", "ab").lastindex, re.match(r"(a)(b)", "ab").lastindex)
print(re.match(r"(x)?(?(1)a|b)", "xa").group(), re.match(r"(x)?(?(1)a|b)", "b").group(), re.match(r"(?P<q>')?\w+(?(q)')", "'ab'").group(), re.match(r"(?P<q>')?\w+(?(q)')", "ab").group())
m = re.match(r"(\d+)-(\d+)", "10-20")
a, b = list(map(int, m.groups()))
print(a + b, [m.group(i) for i in range(3)], len(m.groups()), bool(m), m.string, m.re.groups)
for bad in (lambda: m.group(3), lambda: m.group("nope"), lambda: m[5], lambda: re.sub(r"(a)", r"\2", "a"), lambda: m.start("x")):
    try:
        print(bad())
    except (IndexError, re.error) as e:
        print(type(e).__name__, e)
print(repr(m), repr(re.match("a", "a")), repr(re.compile("x", re.I | re.M)), repr(re.compile(b"b")), re.compile("a") == re.compile("a"), re.compile("a") is re.compile("a"))

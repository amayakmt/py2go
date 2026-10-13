import re
m = re.match(rb"(\w+) (\d+)", b"abc 123 xyz")
print(m, m.group(), m.groups(), m.span(), type(m.group()).__name__, re.findall(rb"\d", b"a1b2"), re.sub(rb"\d", b"#", b"a1b2"), re.split(rb",", b"a,b"), re.escape(b"a.b"), re.search(rb"\xff", b"\x00\xff") is not None)
print(re.findall(rb"[\x80-\xff]+", b"ab\xc3\xa9cd"), re.sub(rb"(?P<x>a)", rb"\g<x>\g<x>", b"banana"), re.match(rb"(?i)abc", b"ABC").group(), re.fullmatch(rb".+", b"ab\n"), re.compile(rb"a").flags, re.compile(rb"a").pattern)
print([m.group() for m in re.finditer(rb"\w+", b"one two")], re.compile(rb"x*").split(b"axb"), re.subn(rb"a", lambda m: b"A", b"aa"))

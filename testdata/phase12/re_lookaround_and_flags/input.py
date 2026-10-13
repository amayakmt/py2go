import re
print(re.findall(r"\w+(?=,)", "a, b, c"), re.findall(r"(?<=\$)\d+", "cost $5 and $10"), re.findall(r"(?<!\$)\b\d+", "cost $5 and 10"), re.findall(r"\w+(?!\d)\b", "ab1 cd"), re.sub(r"(?<=\d)(?=(\d{3})+$)", ",", "1234567"))
print(re.findall(r"(?<=ab)c|(?<=de)f", "abc def"), re.search(r"^(?=.*\d)(?=.*[a-z]).{4,}$", "abc1") is not None, re.search(r"^(?=.*\d)(?=.*[a-z]).{4,}$", "abcd"))
print(re.findall("hello", "Hello HELLO hello", re.I), re.findall(r"(?i)hello", "Hello HELLO"), re.findall(r"^\w", "a\nb\nc", re.M), re.findall(r"\w$", "a\nb\nc", re.M), re.findall(r"^\w", "a\nb"), re.search(r"a.b", "a\nb"), re.search(r"a.b", "a\nb", re.S).group() == "a\nb")
print(re.findall(r"""
    (\d+)   # digits
    \s*     # spaces
    ([a-z]+)  # unit
""", "10 kg 20m", re.X), re.findall(r"(?x) a b # comment\n c", "abc"), re.match(r"(?i:a)b", "Ab").group(), re.match(r"(?i:a)b", "AB"), re.match(r"a(?i)b", "aB") is not None if False else "skip", re.findall(r"(?s:.)", "\n"), re.match(r"(?-i:a)", "A", re.I))
print(re.findall(r"é", "É é", re.I), re.findall(r"\w+", "héllo wörld"), re.findall(r"\w+", "héllo wörld", re.A), re.findall(r"\d+", "٣٤ 12"), re.findall(r"\d+", "٣٤ 12", re.A), re.findall(r"\s", "a b"), re.findall(r"\s", "a b", re.A), re.findall("[à-ü]+", "àéü x"))
print(re.I, re.IGNORECASE, re.M | re.S, int(re.X), re.A | re.I, re.compile("a", re.I).flags, re.compile("a").flags, re.compile(b"a").flags)

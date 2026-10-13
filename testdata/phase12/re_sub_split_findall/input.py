import re
print(re.sub(r"a", "b", "banana"), re.sub(r"a", "b", "banana", count=2), re.sub(r"(?P<w>\w+)", r"<\g<w>>", "hi there"), re.sub(r"\d", lambda m: str(int(m.group()) * 2), "a1b2c3"), re.sub(r"^", "> ", "a\nb", flags=re.M), re.sub("", "-", "abc"), re.sub("x*", "-", "abxd"))
print(re.sub(r"(a)(b)?", r"[\1|\2]", "ab a"), re.sub(r"\s+", "", " a b\tc\n"), re.sub(r"a|(b)", r"\1", "ab"), re.sub("b", r"\\", "abc"), re.sub("b", "\\n", "abc"), re.sub(r"(.)", r"\1\1", "abc"), re.sub("a", "\\g<0>\\g<0>", "a"), re.sub(r"\bcat\b", "dog", "cat concat cat"))
print(re.subn(r"a", "b", "banana"), re.subn(r"z", "b", "banana"), re.subn(r"\w+", lambda m: m.group()[::-1], "ab cd"))
print(re.split(r",", "a,b,,c"), re.split(r",", "a,b,,c", maxsplit=2), re.split(r"(,)", "a,b"), re.split(r"\s*", "a b"), re.split(r"(\s)(\w)?", "a b"), re.split("x", "abc"), re.split("", "abc"), re.split(r"\d", "a1b2c3"), re.split(r"[,;]", ""), re.split(r"(?=b)", "abab"))
print(re.findall(r"\w+", "one two  three"), re.findall(r"(\d)(\w)", "1a 2b"), re.findall(r"\d*", "a12b"), re.findall(r"", "ab"), re.findall(r"^", "a\nb", re.M), re.findall(r"(?=(\w\w))", "abc"), re.findall(r"a|", "bab"))
print([(m.start(), m.group()) for m in re.finditer(r"\d+", "a1b22c333")], [m.span() for m in re.finditer(r"", "ab")], [m.group(1) for m in re.finditer(r"<(\w+)>", "<a><b>")], list(re.finditer("z", "abc")))
words = "the quick brown fox"
print(re.sub(r"(\w+)", lambda m: m.group(1).capitalize(), words), re.sub(r"(?<=\w)(?=\w)", "-", "abc"), " ".join(re.findall(r"\b\w", words)).upper(), re.sub(r"[aeiou]", "", words), len(re.findall(r"o", words)))
for bad in (lambda: re.sub("a", "\\q", "a"), lambda: re.sub("a", 5, "a"), lambda: re.sub("a", "b", 5), lambda: re.sub("a", "\\g<9>", "a"), lambda: re.sub("a", "\\g<x>", "a"), lambda: re.sub(b"a", "b", b"a"), lambda: re.sub("a", b"b", "a")):
    try:
        print(bad())
    except (re.error, TypeError, IndexError) as e:
        print(type(e).__name__, e)

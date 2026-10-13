import re
cases = [(r"a", "a"), (r"a", "b"), (r"a*", "aaab"), (r"a+", "baaa"), (r"a?b", "ab"), (r"a{2}", "aaa"), (r"a{2,3}", "aaaa"), (r"a{,2}", "aaa"), (r"a{2,}", "aaaaa"), (r"a{0}", "a"), (r".", "\n"), (r"..", "ab"),
         (r"^a", "a"), (r"a$", "a\n"), (r"\Aa\Z", "a"), (r"\Aa\Z", "a\n"), (r"[abc]+", "xxabcabcxx"), (r"[^abc]+", "abcxyzabc"), (r"[a-c]", "b"), (r"[a-]", "-"), (r"[-a]", "-"), (r"[]a]", "]"), (r"[\]]", "]"), (r"[\w.]+", "foo.bar baz"),
         (r"\d+", "abc123"), (r"\D+", "123abc"), (r"\w+", "  hello_1 "), (r"\W+", "ab-- cd"), (r"\s+", "ab \t\ncd"), (r"\S+", "  ab  "), (r"\bfoo\b", "a foo b"), (r"\Bfoo", "afoo"), (r"a|b|c", "xcx"), (r"(a|b)+", "abba!"), (r"(?:ab)+", "ababx"),
         (r"a.*b", "axxbxxb"), (r"a.*?b", "axxbxxb"), (r"a+?", "aaa"), (r"a*?b", "aab"), (r"x{2,3}?", "xxxx"), (r"a++b", "aaab"), (r"(?>a+)b", "aaab"), (r"a*+a", "aaa"), (r"\x41B\103", "ABC"), (r"\.", "a.b"), (r"\\", "a\\b"), (r"\n", "\n"), (r"\t\r", "\t\r")]
for pat, text in cases:
    m = re.search(pat, text)
    print(repr(pat), repr(text), m.span() if m else None, repr(m.group()) if m else None)
for pat, text in ((r"a", "ab"), (r"b", "ab"), (r"ab", "ab"), (r"a|ab", "ab"), (r"(a|ab)(c|bcd)", "abcd")):
    print(pat, re.match(pat, text), re.fullmatch(pat, text))

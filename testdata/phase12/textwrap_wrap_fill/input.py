import textwrap
text = "The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs."
for w in (10, 20, 35, 80):
    lines = textwrap.wrap(text, w)
    print(w, len(lines), max(len(l) for l in lines), lines[:2])
print(textwrap.fill(text, 30))
print(textwrap.fill(text, 30, initial_indent="  * ", subsequent_indent="    "))
print(textwrap.wrap("a-very-long-hyphenated-word and normal", 10), textwrap.wrap("abcdefghijklmnop", 5), textwrap.wrap("abcdefghijklmnop", 5, break_long_words=False), textwrap.wrap("", 10), textwrap.wrap("   ", 10), textwrap.wrap("one\ntwo\tthree", 20))
print(textwrap.shorten("Hello world, this is a long sentence", 20), textwrap.shorten("short", 20), textwrap.shorten("  multiple   spaces here  ", 12), textwrap.shorten("Hello world", 8, placeholder="..."))
print(repr(textwrap.dedent("    line1\n      line2\n    line3\n")), repr(textwrap.dedent("  a\n\n  b\n")), repr(textwrap.dedent("a\n  b")), repr(textwrap.dedent("\tx\n\t y")))
print(repr(textwrap.indent("a\nb\n\nc", "  ")), repr(textwrap.indent("a\n  \nb", "> ")), repr(textwrap.indent("x\ny", "+", lambda l: l.startswith("y"))))

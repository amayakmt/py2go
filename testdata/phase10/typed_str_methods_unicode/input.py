def clean(s: str) -> str:
    return " ".join(s.split()).strip().lower()

def title_words(s: str) -> list[str]:
    return [w.capitalize() for w in s.split()]

def count_vowels(s: str) -> int:
    return sum(1 for c in s.lower() if c in "aeiouáéíóúü")

text: str = "  Héllo   WÖRLD  straße  ǅungla  "
print(repr(clean(text)), title_words(text), count_vowels(text), text.upper(), text.casefold().strip(), text.swapcase(), text.title(), text.isascii(), len(text), len(text.strip()), text.strip().encode("utf-8"), len(text.encode()), text.find("WÖ"), text.count("l"))
s: str = "ÀÉÎõü"
print(s.lower(), s.upper(), s.isalpha(), s.islower(), s.isupper(), s[::-1], s[1:3], s.center(9, "·"), s.ljust(7, "-"), s.rjust(7, "-"), ord(s[0]), chr(ord(s[0]) + 32), [hex(ord(c)) for c in s], sorted(s), max(s), min(s), s.encode("utf-16-le").hex(), s.encode("latin-1", "replace"))
t: str = "a,b;c d"
import_free = t.replace(";", ",").replace(" ", ",").split(",")
print(import_free, ",".join(reversed(import_free)), "|".join(sorted(import_free, reverse=True)), t.partition(";"), t.rpartition(","), t.splitlines(), t.translate(str.maketrans(",;", "  ")), t.zfill(10), t.expandtabs(), t.startswith(("a", "b")), t.endswith("d"), "%s|%5s|%-5s|" % (t[:1], t[:2], t[:3]), f"{t!r:>12}|{t:>12}|{t:^12}|{t:*<12}|")

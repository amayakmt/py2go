import unicodedata as ud

chars = ["a", "A", "5", " ", "é", "ǅ", "€", "中", "한", "٣", "½", "́", "‏", "😀", "\x00", "ª", "Ⅷ", "\U0001F1E6", "ﬁ", "­"]
for c in chars:
    print(repr(c), ud.category(c), ud.bidirectional(c), ud.combining(c), ud.east_asian_width(c), ud.mirrored(c), ud.decomposition(c), ud.decimal(c, None), ud.digit(c, None), ud.numeric(c, None), ud.name(c, None))
print(ud.name("a"), ud.name("中"), ud.name("한"), ud.name("😀"), ud.lookup("LATIN SMALL LETTER A"), ud.lookup("greek small letter alpha"), ud.lookup("CJK UNIFIED IDEOGRAPH-4E2D"), ud.lookup("HANGUL SYLLABLE HAN"), ud.lookup("GRINNING FACE"))
for text in ["é", "é", "Å", "Å", "ﬁ", "①", "한글", "한", "ạ̇", "ạ̇", "ǆ", "ｱ", "ẛ̣", "Å", "x̧́y"]:
    print(repr(text), [ud.normalize(f, text).encode("unicode_escape").decode() for f in ("NFC", "NFD", "NFKC", "NFKD")], ud.is_normalized("NFC", text), ud.is_normalized("NFKD", text))
for call in (lambda: ud.name("\x00"), lambda: ud.decimal("a"), lambda: ud.digit("a"), lambda: ud.numeric("a"), lambda: ud.lookup("NO SUCH NAME"), lambda: ud.category("ab"), lambda: ud.category(""), lambda: ud.normalize("XYZ", "a"), lambda: ud.category(5)):
    try:
        print(call())
    except (ValueError, KeyError, TypeError) as e:
        print(type(e).__name__, e)
print(ud.unidata_version, sum(1 for i in range(0x110000) if ud.category(chr(i)) == "Lu") if False else "")
counts = {}
for i in [j for j in range(0, 0x30000, 7) if not 0xD800 <= j <= 0xDFFF]:
    c = ud.category(chr(i))
    counts[c] = counts.get(c, 0) + 1
print(sorted(counts.items()))
s = "Ångström Åland ﬁancé 한국어 é̂"
print(ud.normalize("NFC", s) == s, len(ud.normalize("NFD", s)), len(ud.normalize("NFKC", s)), sorted(set(ud.normalize("NFKD", s)))[:5])

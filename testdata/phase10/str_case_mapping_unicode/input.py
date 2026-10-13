samples = ["hello", "HELLO", "Straße", "ǅ", "ß", "ﬁ", "İstanbul", "ΑΣ", "ΑΣΑ", "ὈΔΥΣΣΕΎΣ", "ŉ", "ǰ", "Ǆ", "ǆ", "ꙋ", "ᾳ", "ᾼ", "ΐ", "ẞ", "Ⅻ", "ⓐ", "σς", "Σ", "ﬃ", "µ", "ÿ", "ĳ", "ŀ", "ᲀ"]
for s in samples:
    print(repr(s), repr(s.upper()), repr(s.lower()), repr(s.title()), repr(s.capitalize()), repr(s.swapcase()), repr(s.casefold()), len(s.upper()), len(s.lower()), s.isupper(), s.islower(), s.istitle())
print("ǆ".title(), "ǉ".capitalize(), "hello wORLD".title(), "they're bill's".title(), "x1y2z".title(), "ÀÉÎ".lower(), "àéî".upper(), "ß".upper() == "SS", "SS".lower() == "ss", "ß".casefold() == "ss")
print(["a".isupper(), "A".isupper(), "A1".isupper(), "1".isupper(), "".isupper(), "aB".isupper(), "ab".islower(), "a b".islower(), " ".islower(), "Ab Cd".istitle(), "AB".istitle(), "Ab cd".istitle(), "".istitle(), "ǅ".istitle()])
print("i̇".lower() == "i̇", "İ".lower(), len("İ".lower()), "ΣΑΣ".lower(), "ΑΣ ".lower(), "Σ".lower(), "ΑΣ.".lower(), "ASΣ".lower())

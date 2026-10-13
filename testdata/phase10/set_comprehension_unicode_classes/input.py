text = "Hello Wörld! ΑΒΓ αβγ 日本語 😀😀 123 ٣٤٥ ½ ² _x-y"
classes = {}
for ch in text:
    key = ("alpha" if ch.isalpha() else "") + ("digit" if ch.isdigit() else "") + ("decimal" if ch.isdecimal() else "") + ("numeric" if ch.isnumeric() else "") + ("space" if ch.isspace() else "") + ("upper" if ch.isupper() else "") + ("lower" if ch.islower() else "") + ("ascii" if ch.isascii() else "") + ("print" if ch.isprintable() else "") + ("ident" if ch.isidentifier() else "")
    classes.setdefault(key, set()).add(ch)
for k in sorted(classes):
    print(k or "(none)", "".join(sorted(classes[k])))
print({c.upper() for c in text if c.isalpha()} >= set("HELLO"), len({c.casefold() for c in text if c.isalpha()}), sorted({ord(c) // 100 for c in text if not c.isascii()}), sorted({c for c in text if c.isspace()}), len({c for c in text}))

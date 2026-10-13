text = """Der Fluß fließt; the river flows. Ça va? Ça va très bien! ΑΣ ΑΣΑ Ελλάδα ελλάδα.
日本語のテキスト and ünïcode wörds, naïve café CAFÉ café. ǅ ǆ Ǆ\tTabs\tand   spaces."""
import_words = []
word = ""
for ch in text + " ":
    if ch.isalpha():
        word += ch
    else:
        if word:
            import_words.append(word)
        word = ""
freq = {}
for w in import_words:
    k = w.casefold()
    freq[k] = freq.get(k, 0) + 1
top = sorted(freq.items(), key=lambda kv: (-kv[1], kv[0]))
print(top[:8])
print(len(import_words), len(freq), max(import_words, key=len), min(import_words), sorted(import_words, key=str.lower)[:6])
print([w.title() for w in import_words[:6]], [w.upper() for w in import_words[:4]], [w.swapcase() for w in import_words[10:14]], [len(w.encode()) for w in import_words[:8]])
lines = text.splitlines()
print([len(l) for l in lines], [len(l.split()) for l in lines], [l.count("ç") + l.count("Ç") for l in lines], text.count("a"), text.lower().count("a"), text.casefold().count("ss"), text.find("café"), text.rfind("café"), text.lower().find("café"))
print(sorted(set(c for c in text if not c.isascii()))[:12], sum(c.isspace() for c in text), sum(c.isupper() for c in text), sum(c.isdigit() for c in text), "".join(sorted(set(text.lower()) & set("aeiouyäöüéèàç"))))

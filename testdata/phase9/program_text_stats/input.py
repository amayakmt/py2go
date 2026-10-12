text = """The quick brown fox jumps over the lazy dog.
The dog barks; the fox runs away!
Quick, quick, said the fox."""

def words(s):
    out = []
    for line in s.split("\n"):
        for w in line.replace(",", " ").replace(".", " ").replace(";", " ").replace("!", " ").split():
            out.append(w.lower())
    return out

ws = words(text)
freq = {}
for w in ws:
    freq[w] = freq.get(w, 0) + 1
top = sorted(freq.items(), key=lambda kv: (-kv[1], kv[0]))[:5]
print(len(ws), len(freq), top)
longest = max(ws, key=len)
print(longest, len(longest), [w for w in set(ws) if len(w) == len(longest)] == [longest] or sorted(w for w in set(ws) if len(w) == len(longest)))
by_len = {}
for w in ws:
    by_len.setdefault(len(w), set()).add(w)
print({k: sorted(v) for k, v in sorted(by_len.items())})
lines = text.split("\n")
print([len(l) for l in lines], [l.split()[0] for l in lines], text.count("the"), text.lower().count("the"))
print(" ".join(w.capitalize() for w in ws[:4]), "-".join(sorted(set(w[0] for w in ws))))
print(text.find("fox"), text.rfind("fox"), text.index("dog"), text.startswith("The"), text.endswith("."))

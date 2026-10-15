words = ["alpha", "beta", "gamma", "delta", "epsilon"] * 40000
counts = {}
for w in words:
    key = w.upper() + str(len(w))
    counts[key] = counts.get(key, 0) + 1
print(sorted(counts.items()))
text = " ".join(words)
print(len(text.split()), text.count("gamma"))
items = sorted(words, key=lambda s: (len(s), s))
print(items[0], items[-1])

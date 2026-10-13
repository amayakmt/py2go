words = [b"apple", b"banana", b"cherry", b"date"]
d = {w: len(w) for w in words}
print(d, {w[:1]: w for w in words}, {w.decode(): w.upper() for w in words}, {k: v for k, v in d.items() if v > 4}, dict(zip(words, range(4))), {b"x": 1, b"x": 2}, d[b"date"], b"fig" in d, list(d), sorted(d, key=len), max(d, key=d.get))
inv = {}
for w in words:
    inv.setdefault(len(w), []).append(w)
print(inv, {k: b"+".join(v) for k, v in inv.items()}, sorted(inv.items()), sum(map(len, words)), b"".join(sorted(words)), [w[::-1] for w in words], [w.title() for w in words], {bytes([c]) for w in words for c in w} >= {b"a", b"e"}, len({bytes([c]) for w in words for c in w}))
t = {(1, b"a"): "x", (2, b"b"): "y", (1, b"a", 2 ** 70): "z"}
print(t, t[(1, b"a")], t.get((1, b"a", 2 ** 70)), (2, b"b") in t, sorted(t, key=len))
counts = {}
for ch in b"mississippi":
    counts[bytes([ch])] = counts.get(bytes([ch]), 0) + 1
print(counts, sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])), {chr(k[0]): v for k, v in counts.items()})

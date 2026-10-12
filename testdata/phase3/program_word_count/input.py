text: str = "the quick brown fox jumps over the lazy dog the end"
words: list[str] = text.split()
seen: list[str] = []
counts: list[int] = []
for w in words:
    if w in seen:
        counts[seen.index(w)] += 1
    else:
        seen.append(w)
        counts.append(1)
for w, c in zip(seen, counts):
    print(f"{w:>6}: {c}")
best: int = max(counts)
print(best, seen[counts.index(best)])

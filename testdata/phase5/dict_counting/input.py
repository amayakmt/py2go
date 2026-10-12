text: str = "the cat and the hat and the bat"
counts: dict[str, int] = {}
for word in text.split():
    counts[word] = counts.get(word, 0) + 1
print(counts)
best: str = ""
for word in counts:
    if best == "" or counts[word] > counts[best]:
        best = word
print(best, counts[best])
letters: dict[str, int] = {}
for ch in "mississippi":
    if ch in letters:
        letters[ch] += 1
    else:
        letters[ch] = 1
print(letters)
for ch in sorted(letters):
    print(ch, letters[ch])

words: list[str] = ["pear", "fig", "banana", "kiwi", "plum"]
words.sort(key=len)
print(words)
words.sort(key=lambda w: w[1], reverse=True)
print(words)
print(min(words, key=len), max(words, key=len))
print(min("bb", "a", "ccc", key=len), max(3, -5, 4, key=abs))
print(min([("a", 3), ("b", 1)], key=lambda p: p[1]))
print(max(range(5), key=lambda i: (i % 3, i)))
calls: list[int] = [0]


def key(w: str, calls: list[int]) -> int:
    calls[0] += 1
    return len(w)


big: list[str] = ["ccc", "a", "bb", "dd"] * 3
big.sort(key=lambda w: key(w, calls))
print(big, calls[0])
print(max(["aa", "bb", "cc"], key=len), min(["aa", "bb", "cc"], key=len))

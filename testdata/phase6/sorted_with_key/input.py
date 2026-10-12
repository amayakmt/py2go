words: list[str] = ["banana", "kiwi", "apple", "fig", "cherry"]
print(sorted(words, key=len))
print(sorted(words, key=lambda w: w[-1]))
print(sorted(words, key=lambda w: (len(w), w)))
print(sorted(words, key=lambda w: -len(w)))
print(sorted(words, key=len, reverse=True))
print(sorted(words, key=lambda w: w.lower()))
nums: list[int] = [-3, 1, -2, 4, 0]
print(sorted(nums, key=abs), sorted(nums, key=abs, reverse=True))
print(sorted(nums, key=lambda n: n * n))
pairs: list[tuple[str, int]] = [("b", 2), ("a", 2), ("c", 1)]
print(sorted(pairs, key=lambda p: p[1]))
print(sorted(pairs, key=lambda p: (p[1], p[0])))
print(sorted(pairs, key=lambda p: p[1], reverse=True))
d: dict[str, int] = {"x": 3, "y": 1, "z": 2}
print(sorted(d, key=lambda k: d[k]), sorted(d.items(), key=lambda kv: kv[1]))
print(sorted(range(5), key=lambda i: -i))

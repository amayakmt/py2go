print(set([3, 1, 3, 2]))
print(sorted(set("mississippi")))
print(set(range(5)))
print(set(x * x for x in range(6)))
d: dict[int, str] = {5: "a", 3: "b", 8: "c"}
print(set(d), set(d.values()) == {"a", "b", "c"})
print(len(set([(1, 2), (1, 2), (2, 1)])))
words: list[str] = ["b", "a", "b", "c", "a"]
print(sorted(set(words)), len(set(words)))
uniq: list[int] = []
seen: set[int] = set()
for x in [4, 1, 4, 2, 1, 3]:
    if x not in seen:
        seen.add(x)
        uniq.append(x)
print(uniq)

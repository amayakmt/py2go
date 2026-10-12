s: set[str] = {"pear", "apple", "fig", "apple"}
print(sorted(s), len(s), "fig" in s)
s.add("kiwi")
s.discard("pear")
print(sorted(s))
t: set[str] = {"fig", "plum"}
print(sorted(s | t), sorted(s & t), sorted(s - t), sorted(s ^ t))
for w in sorted(s):
    print(w)

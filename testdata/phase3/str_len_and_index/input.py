s: str = "hello"
print(len(s), s[0], s[4], s[-1], s[-5])
print(len(""), len("a"))
i: int = 2
print(s[i], s[i + 1], s[-i])
for k in range(len(s)):
    print(k, s[k])

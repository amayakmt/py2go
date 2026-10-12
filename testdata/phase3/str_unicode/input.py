s: str = "héllo wörld"
print(len(s), s[1], s[-1], s[7])
print(s.upper(), s.find("w"), s.find("ö"))
k: str = "日本語"
print(len(k), k[1], k[::-1], k[-1])
print("é" in s, "e" in s)
print(s[1:4], s[::-1])
e: str = "😀!"
print(len(e), e[0], e[1])

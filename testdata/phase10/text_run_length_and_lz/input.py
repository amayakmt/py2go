def rle(s):
    out, i = [], 0
    while i < len(s):
        j = i
        while j < len(s) and s[j] == s[i]:
            j += 1
        out.append((s[i], j - i))
        i = j
    return out

def unrle(pairs):
    return "".join(c * n for c, n in pairs)

def lz78(s):
    d, out, cur = {}, [], ""
    for ch in s:
        if cur + ch in d:
            cur += ch
        else:
            out.append((d.get(cur, 0), ch))
            d[cur + ch] = len(d) + 1
            cur = ""
    if cur:
        out.append((d[cur], ""))
    return out

def unlz78(pairs):
    d, out = {0: ""}, []
    for i, (idx, ch) in enumerate(pairs, 1):
        s = d[idx] + ch
        d[i] = s
        out.append(s)
    return "".join(out)

for s in ("", "aaabccddddde", "abababababab", "日日日本本語語語語", "😀😀😀 hello hello", "a" * 50 + "b" * 3, "abracadabra" * 3):
    r = rle(s)
    z = lz78(s)
    print(repr(s[:20]), r[:5], len(r), unrle(r) == s, z[:6], len(z), unlz78(z) == s)
print("".join(f"{n}{c}" for c, n in rle("WWWWWWWWWWWWBWWWWWWWWWWWWBBBWWWWWWWWWWWWWWWWWWWWWWWWBWWWWWWWWWWWWWW")))

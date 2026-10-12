def encode(s):
    if not s:
        return []
    out = []
    prev, count = s[0], 1
    for c in s[1:]:
        if c == prev:
            count += 1
        else:
            out.append((prev, count))
            prev, count = c, 1
    out.append((prev, count))
    return out

def decode(pairs):
    return "".join(c * n for c, n in pairs)

def compress(s):
    return "".join(c + (str(n) if n > 1 else "") for c, n in encode(s))

for s in ("aaabccdddd", "abc", "", "zzzzzzzzzzzz", "aabbaa"):
    e = encode(s)
    print(repr(s), e, compress(s), decode(e) == s)
nums = [1, 1, 2, 3, 3, 3]
print(encode(nums), decode([("x", 2), ("y", 0), ("z", 1)]))
print(sorted(set(compress("aaabbbcccd"))), len(compress("a" * 100)))

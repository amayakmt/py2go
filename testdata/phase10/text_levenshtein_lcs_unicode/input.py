def lev(a, b):
    prev = list(range(len(b) + 1))
    for i, ca in enumerate(a, 1):
        cur = [i]
        for j, cb in enumerate(b, 1):
            cur.append(min(prev[j] + 1, cur[j - 1] + 1, prev[j - 1] + (ca != cb)))
        prev = cur
    return prev[-1]

def lcs(a, b):
    dp = [[""] * (len(b) + 1) for _ in range(len(a) + 1)]
    for i in range(len(a)):
        for j in range(len(b)):
            if a[i] == b[j]:
                dp[i + 1][j + 1] = dp[i][j] + a[i]
            else:
                dp[i + 1][j + 1] = max(dp[i][j + 1], dp[i + 1][j], key=len)
    return dp[-1][-1]

pairs = [("kitten", "sitting"), ("flaw", "lawn"), ("", "abc"), ("naïve", "naive"), ("日本語", "日本人"), ("😀a😀", "a😀b"), ("straße", "strasse"), ("abc", "abc")]
for a, b in pairs:
    print(a, b, lev(a, b), lcs(a, b), lev(a.casefold(), b.casefold()), len(a.encode()), len(b.encode()))
words = "the quick brown fox jumps over the lazy dog".split()
print(sorted(words, key=lambda w: lev(w, "quack"))[:4], min(words, key=lambda w: lev(w, "lazzy")), sum(lev(w, "the") for w in words), [lev("a" * n, "a" * (2 * n)) for n in range(5)])

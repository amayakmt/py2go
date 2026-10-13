ROMAN = [(1000, "M"), (900, "CM"), (500, "D"), (400, "CD"), (100, "C"), (90, "XC"), (50, "L"), (40, "XL"), (10, "X"), (9, "IX"), (5, "V"), (4, "IV"), (1, "I")]

def to_roman(n):
    out = ""
    for v, s in ROMAN:
        while n >= v:
            out += s
            n -= v
    return out

def from_roman(s):
    total, prev = 0, 0
    for ch in reversed(s):
        v = dict((b, a) for a, b in ROMAN)[ch]
        total += v if v >= prev else -v
        prev = max(prev, v)
    return total

print([to_roman(n) for n in (1, 4, 9, 14, 40, 90, 400, 1994, 2024, 3999)], all(from_roman(to_roman(n)) == n for n in range(1, 4000)))
ONES = "zero one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen seventeen eighteen nineteen".split()
TENS = "_ _ twenty thirty forty fifty sixty seventy eighty ninety".split()
SCALES = ["", "thousand", "million", "billion", "trillion", "quadrillion", "quintillion", "sextillion"]

def words(n):
    if n < 0:
        return "minus " + words(-n)
    if n < 20:
        return ONES[n]
    if n < 100:
        return TENS[n // 10] + ("-" + ONES[n % 10] if n % 10 else "")
    if n < 1000:
        return ONES[n // 100] + " hundred" + (" and " + words(n % 100) if n % 100 else "")
    for i in range(len(SCALES) - 1, 0, -1):
        if n >= 1000 ** i:
            q, r = divmod(n, 1000 ** i)
            return words(q) + " " + SCALES[i] + (", " + words(r) if r else "")

for n in (0, 7, 13, 21, 99, 100, 101, 342, 1000, 1001, 12345, 1000000, 2 ** 31, -45, 10 ** 18 + 1, 2 ** 64, 123456789012345678901):
    print(n, words(n))

ROMAN = [(1000, "M"), (900, "CM"), (500, "D"), (400, "CD"), (100, "C"), (90, "XC"),
         (50, "L"), (40, "XL"), (10, "X"), (9, "IX"), (5, "V"), (4, "IV"), (1, "I")]

def to_roman(n):
    out = ""
    for value, symbol in ROMAN:
        while n >= value:
            out += symbol
            n -= value
    return out

def from_roman(s):
    total = 0
    for i, c in enumerate(s):
        v = dict((sym, val) for val, sym in ROMAN)[c]
        if i + 1 < len(s) and dict((sym, val) for val, sym in ROMAN)[s[i + 1]] > v:
            total -= v
        else:
            total += v
    return total

for n in (1, 4, 9, 14, 40, 90, 400, 1994, 2024, 3999):
    r = to_roman(n)
    print(n, r, from_roman(r) == n)

def to_base(n, base):
    digits = "0123456789abcdef"
    if n == 0:
        return "0"
    out = ""
    neg = n < 0
    n = abs(n)
    while n:
        out = digits[n % base] + out
        n //= base
    return "-" + out if neg else out

print([to_base(255, b) for b in (2, 8, 10, 16)], to_base(-10, 2), to_base(0, 7))
print(bin(10), oct(64), hex(255), int("ff", 16) if False else 255, chr(65), ord("a"), [chr(ord("a") + i) for i in range(5)])

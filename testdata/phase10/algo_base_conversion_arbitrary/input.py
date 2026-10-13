DIGITS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

def to_base(n, b):
    if n == 0:
        return "0"
    sign = "-" if n < 0 else ""
    n = abs(n)
    out = []
    while n:
        n, r = divmod(n, b)
        out.append(DIGITS[r])
    return sign + "".join(reversed(out))

def from_base(s, b):
    neg = s.startswith("-")
    n = 0
    for ch in s.lstrip("-"):
        n = n * b + DIGITS.index(ch)
    return -n if neg else n

for n in (0, 1, -1, 255, 2 ** 64, -(2 ** 70) + 3, 10 ** 40):
    for b in (2, 8, 16, 36, 62):
        s = to_base(n, b)
        assert from_base(s, b) == n
    print(n, to_base(n, 2)[:20], to_base(n, 16), to_base(n, 36), to_base(n, 62), int(to_base(n, 36), 36) == n, int(to_base(n, 16), 16) == n)
print(int("zz", 36), int("-ff", 16), int("0b101", 0), int("0o17", 0), int("0x_ff", 0), int("1_000", 10), int(" 12 ", 10), int("z" * 30, 36))
for bad in (("12", 2), ("g", 16), ("", 10), ("1__0", 10), ("0x", 16), (" ", 10), ("1.5", 10), ("5", 1), ("5", 37)):
    try:
        print(int(bad[0], bad[1]))
    except ValueError as e:
        print(e)
print(bin(2 ** 70), oct(2 ** 70), hex(-2 ** 70), format(2 ** 70, "b"), format(-2 ** 70, "x"), format(2 ** 70, "#o"), format(2 ** 70, ",d"), f"{2 ** 70:_x}")

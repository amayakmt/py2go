good = ["0", "-0", "+7", "007", "1_000", "  42  ", "\n-5\t", "٣٤", "１２３", "123456789012345678901234567890", "-" + "9" * 40, "0" * 50, "1" + "0" * 100]
for s in good:
    v = int(s)
    print(repr(s)[:30], v if abs(v) < 10 ** 35 else len(str(v)), type(v).__name__)
bad = ["", " ", "1.5", "1e3", "0x10", "1 2", "--1", "+-1", "1_", "_1", "1__0", "0b1", "١٢a", "½", "²", "1,000", "- 1", "٣.٤", "\x00", "12\x00"]
for s in bad:
    try:
        print(repr(s), int(s))
    except ValueError as e:
        print(repr(s), "ValueError", e)
for s, b in (("ff", 16), ("FF", 16), ("0xff", 16), ("0XFF", 0), ("-0b101", 0), ("0o777", 8), ("777", 8), ("zz", 36), ("Zz", 36), ("1_0", 2), ("102", 2), ("0b", 0), ("1" * 70, 2), ("9" * 30, 10), ("-" + "f" * 30, 16)):
    try:
        v = int(s, b)
        print(s[:12], b, v if abs(v) < 10 ** 20 else len(str(v)))
    except ValueError as e:
        print(s[:12], b, "ValueError", e)

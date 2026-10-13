n = 2 ** 100 + 12345
print(n, str(n), repr(-n), f"{n}", "%d" % n, "%s" % n, "%x" % n, "{}".format(n))
print(f"{n:,}", f"{n:_}", f"{n:x}", f"{n:X}", f"{n:o}", f"{n:b}"[:20], f"{n:#x}", f"{n:e}", f"{n:.3e}", f"{n:40d}|", f"{n:<40d}|", f"{-n:^45d}|", f"{n:+d}")
print(f"{n:0>40}", f"{n:*^40}", f"{2**70:020,}", f"{-2**70:+,}", format(n, "x"), format(2 ** 64, "_x"))
print(int("123456789012345678901234567890"), int(" +42 "), int("-1_000"), int("ff", 16), int("0xFF", 16), int("0b101", 0), int("z", 36), int("12345678901234567890123", 8) if False else 0)
print(int("1" * 30), int("-" + "9" * 25), int("0" * 50 + "7"), int(1e25), int(123456789.987), int(-0.9))
for bad in ("1.5", "", "abc", "12 3", "1__0", "0x1G", "١٢٣x"):
    try:
        print(int(bad))
    except ValueError as e:
        print("ValueError", e)
print(str(10 ** 4000)[:10], len(str(10 ** 4000)), len(str(10 ** 4299)))
try:
    str(10 ** 4300)
except ValueError as e:
    print("ValueError", e)
try:
    int("1" * 4301)
except ValueError as e:
    print("ValueError", e)

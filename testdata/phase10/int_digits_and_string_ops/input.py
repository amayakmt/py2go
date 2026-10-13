n = 3 ** 200
s = str(n)
print(len(s), s[:10], s[-10:], sum(int(c) for c in s), s.count("7"), int(s[::-1]) % 1000, s == str(int(s)), n == int(s))
print(int("9" * 100) + 1 == 10 ** 100, int("1" + "0" * 50) == 10 ** 50, len(str(2 ** 10000)), str(2 ** 10000)[:15])
digits = []
m = 2 ** 128
while m:
    digits.append(m % 10)
    m //= 10
print(digits[::-1] == [int(c) for c in str(2 ** 128)], sum(digits))
print("%d|%5d|%-30d|%030d|%+d" % (2 ** 70, 5, 2 ** 70, 2 ** 70, 2 ** 70), "{:>30}|{:<30}|{:^30}|".format(2 ** 70, 2 ** 70, 2 ** 70))
print(repr(2 ** 64), str(-2 ** 64), f"{2**64!r}", f"{2**64!s:>25}", [2 ** 64, -2 ** 64], (2 ** 64,), {2 ** 64: 2 ** 65})
print(max(str(2 ** 70), str(3 ** 40)), sorted(map(str, [10 ** 20, 5, 2 ** 70])), int(str(2 ** 70)) == 2 ** 70, "".join(map(str, range(30)))[-12:])

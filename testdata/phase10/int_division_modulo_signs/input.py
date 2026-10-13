vals = [7, -7, 2 ** 70 + 3, -(2 ** 70 + 3), 0, 1, -1]
divs = [2, -2, 3, -3, 2 ** 65, -(2 ** 65), 1, -1]
for a in vals:
    for b in divs:
        q, r = divmod(a, b)
        assert q * b + r == a and (r == 0 or (r < 0) == (b < 0)) and abs(r) < abs(b)
        print(a, b, a // b, a % b, q, r)
for a in vals:
    print(a / 3, a / -7, a / 2 ** 65 if a else 0.0)

vals = [0, 1, -1, 5, -5, 255, -256, 2 ** 63 - 1, -2 ** 63, 2 ** 64, -2 ** 64 - 1, 2 ** 100 + 12345, -2 ** 100 - 12345]
for a in vals:
    print(a, ~a, -a, a >> 1, a >> 70, a << 1, a << 65, a & 0xFF, a | 0xFF, a ^ 0xFF, a & -a, a.bit_length())
for a in vals[:7]:
    for b in vals[7:]:
        print(a & b, a | b, a ^ b)
print(bin(5), bin(-5), bin(2 ** 70), oct(64), oct(-64), hex(255), hex(-255), hex(2 ** 70), bin(0), hex(0))

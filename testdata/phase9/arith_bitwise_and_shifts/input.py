def bits(a, b):
    return a & b, a | b, a ^ b, ~a, a << 2, a >> 1

print(bits(12, 10))
print(bits(-5, 3))
print(bits(True, 3))
print(1 << 40, 255 >> 4, -16 >> 2, 5 & 3 | 8, 6 ^ 3 ^ 3)
flags = 0
for bit in (1, 4, 6):
    flags |= 1 << bit
print(flags, bin(flags) if False else flags, flags & (1 << 4) != 0, flags & (1 << 5) != 0)
x = 10
x <<= 2
x |= 3
x ^= 1
x &= 0xFF
x >>= 1
print(x)

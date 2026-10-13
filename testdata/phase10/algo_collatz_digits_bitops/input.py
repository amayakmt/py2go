def collatz_len(n):
    steps = 0
    while n != 1:
        n = n // 2 if n % 2 == 0 else 3 * n + 1
        steps += 1
    return steps

best = max(range(1, 3000), key=collatz_len)
print(best, collatz_len(best), collatz_len(2 ** 100 + 1), collatz_len(27), collatz_len(97))
def digit_sum(n, base=10):
    s = 0
    while n:
        n, r = divmod(n, base)
        s += r
    return s
print(digit_sum(2 ** 100), digit_sum(3 ** 200), digit_sum(2 ** 64 - 1, 2), digit_sum(10 ** 30 - 1, 16), bin(2 ** 70 - 1).count("1"))
def popcount(n):
    c = 0
    while n:
        n &= n - 1
        c += 1
    return c
def reverse_bits(n, width):
    r = 0
    for _ in range(width):
        r = (r << 1) | (n & 1)
        n >>= 1
    return r
print(popcount(2 ** 100 - 1), popcount(0xF0F0F0F0F0F0F0F0F0), reverse_bits(0b1011, 4), reverse_bits(1, 100) == 2 ** 99, bin(reverse_bits(0xDEADBEEF, 32)))
x = 0x123456789ABCDEF0123456789ABCDEF
print(hex(x ^ (x >> 7)), hex((x << 3) & ((1 << 128) - 1)), (x >> 64) & 0xFFFF, x.bit_length(), hex(~x & 0xFFFFFFFF), -x >> 100)

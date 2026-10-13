def gray(n):
    return n ^ (n >> 1)

def ungray(g):
    n = 0
    while g:
        n ^= g
        g >>= 1
    return n

print([gray(i) for i in range(16)], [bin(gray(i))[2:].zfill(4) for i in range(8)], all(ungray(gray(i)) == i for i in range(1000)), ungray(gray(2 ** 90 + 12345)) == 2 ** 90 + 12345)
def lowbit(n):
    return n & -n
def is_pow2(n):
    return n > 0 and n & (n - 1) == 0
print([lowbit(i) for i in (1, 2, 6, 12, 40, 2 ** 70, -8, -12)], [is_pow2(2 ** k) for k in (0, 10, 64, 200)], [is_pow2(2 ** k + 1) for k in (0, 1, 64)], is_pow2(0), is_pow2(-4))
def subsets(mask):
    s = mask
    out = []
    while s:
        out.append(s)
        s = (s - 1) & mask
    out.append(0)
    return out
print(subsets(0b1011), len(subsets(2 ** 10 - 1)), len(subsets(0b1010101010101010101010101010101010101010)))
def next_perm(x):
    t = (x | (x - 1)) + 1
    return t | ((((t & -t) // (x & -x)) >> 1) - 1)
x = 0b00111
seq = []
for _ in range(8):
    seq.append(bin(x))
    x = next_perm(x)
print(seq)
print(bin(0b101101 & ~0b100), 0xFF ^ 0x0F, -1 & 0xFF, (-1 << 8) | 0xAB, (1 << 200) - (1 << 100), ((1 << 200) - 1).bit_length(), ((1 << 200) - 1 >> 150).bit_count())

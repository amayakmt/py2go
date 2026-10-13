def egcd(a, b):
    if a == 0:
        return b, 0, 1
    g, x, y = egcd(b % a, a)
    return g, y - (b // a) * x, x

def modinv(a, m):
    g, x, _ = egcd(a % m, m)
    if g != 1:
        raise ValueError("no inverse")
    return x % m

p, q = 2 ** 61 - 1, 2 ** 89 - 1
n = p * q
phi = (p - 1) * (q - 1)
e = 65537
d = modinv(e, phi)
print(n, phi % 1000, d % 1000, e * d % phi)
for msg in (42, 123456789, 2 ** 100 + 7, n - 1):
    c = pow(msg, e, n)
    m = pow(c, d, n)
    print(msg % 10 ** 6, c % 10 ** 6, m == msg)
text = "hello, rsa!"
nums = [ord(ch) for ch in text]
enc = [pow(x, e, n) for x in nums]
dec = "".join(chr(pow(y, d, n)) for y in enc)
print(dec, len(enc), enc[0] > 2 ** 100)
try:
    modinv(6, 9)
except ValueError as ex:
    print(ex)
print(pow(7, -1, 40), pow(3, 10 ** 30, 10 ** 9 + 7), pow(2, 2 ** 70, 2 ** 61 - 1))

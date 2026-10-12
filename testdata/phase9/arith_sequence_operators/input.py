def rep(a, n):
    return a * n

print(rep("ab", 3), rep([0], 4), rep((1, 2), 2), rep("x", 0), rep([1], -1))
print(3 * "ab", 2 * [1, 2], 2 * (3,))
def join(a, b):
    return a + b

print(join("x", "y"), join([1], [2, 3]), join((1,), ()), join([], []))
s = "abc"
print(s[0], s[-1], s[1:], s[::-1], s[::2])
t = (1, 2, 3, 4, 5)
print(t[1], t[-2], t[1:3], t[::-2])
l = [10, 20, 30, 40]
print(l[0], l[-1], l[1:3], l[::-1], l[:2] + l[2:])
l[1:3] = [0]
print(l)
del l[0]
print(l, len(l), 20 in [20], "b" in "abc", 3 in (1, 2))

def cut(seq, a=None, b=None, c=None):
    return seq[a:b:c]

print(cut("abcdef", 1, 4), cut("abcdef", None, None, -1), cut("abcdef", -3), cut("abcdef", 10), cut("abcdef", 2, 100))
print(cut([1, 2, 3, 4, 5], 1, None, 2), cut([1, 2, 3], -1, None, -1), cut((1, 2, 3, 4), 1, 3), cut(range(10), 2, 8, 3))
l = list(range(10))
l[2:4] = ["a", "b", "c"]
print(l)
l[::2] = [0] * 6
print(l)
del l[1:3]
print(l, l[-2:], l[:-8])
l[:0] = ["head"]
l[len(l):] = ["tail"]
print(l)
s = "hello world"
print(s[:5], s[6:], s[-5:-1], s[::3], s[::-2], s[3:3] == "")

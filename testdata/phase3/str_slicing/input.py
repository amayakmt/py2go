s: str = "abcdefghij"
print(s[1:3], s[:2], s[-3:], s[::-1], s[::2], s[1::3])
print(s[10:], s[-100:2], s[3:1], s[:100], s[-1:-4:-1], s[8:2:-2])
print(s[:], s[0:0], s[5:5])
a: int = 2
b: int = 6
print(s[a:b], s[:a], s[b:], s[a:b:a])

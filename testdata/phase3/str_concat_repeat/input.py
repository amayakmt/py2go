print("ab" * 3, 3 * "ab", "x" * 0, "x" * -1, "-" * 10)
s: str = ""
for i in range(5):
    s += str(i)
print(s)
t: str = "ab"
t *= 3
print(t)
print("a" + "b" + "c")
n: int = 2
print("ha" * n + "!")

print(1, 2, 3, sep="-")
print("a", "b", sep="")
print("no newline", end="")
print(" ... continued")
print("x", "y", sep=", ", end="!\n")
print(sep="*")
print("tab", "separated", sep="\t")
s: str = "|"
print(1, 2, sep=s, end=s + "\n")
for i in range(3):
    print(i, end=" ")
print()

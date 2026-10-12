zero: int = 0
one: int = 1
empty: str = ""
word: str = "w"
fz: float = 0.0
if zero:
    print("zero truthy")
else:
    print("zero falsy")
if one:
    print("one truthy")
if empty:
    print("empty truthy")
else:
    print("empty falsy")
if word:
    print("word truthy")
if fz:
    print("0.0 truthy")
else:
    print("0.0 falsy")
print(not zero)
print(not one)
print(not empty)
print(not 0.5)
while one:
    one -= 1
    print("looped once")

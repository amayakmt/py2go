if True:
    picked: str = "yes"
for i in range(3):
    last: int = i
while False:
    never: int = 0
try:
    parsed: int = int("12")
except ValueError:
    parsed = -1


def show() -> str:
    return picked + str(last) + str(i) + str(parsed)


print(show(), picked, last, i, parsed)

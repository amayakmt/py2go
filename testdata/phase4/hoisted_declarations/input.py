def parse(s: str) -> int:
    try:
        n: int = int(s)
        doubled: int = n * 2
    except ValueError:
        n = 0
        doubled = -1
    return n + doubled


print(parse("21"), parse("x"))

label: str = "start"
try:
    value: float = float("2.5")
    label = "parsed"
except ValueError:
    value = 0.0
print(label, value)

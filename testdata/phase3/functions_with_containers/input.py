def evens(xs: list[int]) -> list[int]:
    out: list[int] = []
    for x in xs:
        if x % 2 == 0:
            out.append(x)
    return out


def pair_up(a: list[int], b: list[str]) -> list[tuple[int, str]]:
    out: list[tuple[int, str]] = []
    for x, y in zip(a, b):
        out.append((x, y))
    return out


def describe(t: tuple[str, int]) -> str:
    name, age = t
    return f"{name} is {age}"


def first_word(s: str) -> str:
    parts: list[str] = s.split()
    if not parts:
        return ""
    return parts[0]


print(evens([1, 2, 3, 4, 5, 6]), evens([]))
print(pair_up([1, 2, 3], ["a", "b"]))
print(describe(("Ada", 36)))
print(repr(first_word("  hello world")), repr(first_word("   ")))
for p in pair_up([7, 8], ["x", "y"]):
    print(describe((p[1], p[0])))

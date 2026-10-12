xs: list[int] = [1, 2, 3]
t: tuple[int, str] = (1, "a")
a, b = t
y = xs[1:], xs[::-1], xs[0], "abc".upper()
print(f"{a!r:>5} and {{b}}", sep="-", end="")

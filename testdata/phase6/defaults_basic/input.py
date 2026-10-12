def power(base: int, exp: int = 2) -> int:
    r: int = 1
    for _ in range(exp):
        r *= base
    return r


def label(name: str, prefix: str = "<", suffix: str = ">") -> str:
    return prefix + name + suffix


print(power(3), power(3, 3), power(base=2), power(exp=5, base=2))
print(label("a"), label("a", "["), label("a", "[", "]"), label("a", suffix="!"), label("a", suffix="!", prefix="*"))

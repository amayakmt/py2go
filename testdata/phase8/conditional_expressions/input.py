def sign(n: int) -> str:
    return "neg" if n < 0 else "zero" if n == 0 else "pos"


print([sign(n) for n in [-2, 0, 5]])
x: int = 7
y: float = 1.5 if x > 5 else 2.5
print(y, "big" if x > 100 else "small")
calls: list[str] = []


def noisy(tag: str, v: int) -> int:
    calls.append(tag)
    return v


r: int = noisy("a", 1) if x > 0 else noisy("b", 2)
print(r, calls)
print(max(x, 3) if x else 0, len("abc") if True else 0)
words: list[str] = ["a", "bb"]
print([w.upper() if len(w) > 1 else w for w in words])

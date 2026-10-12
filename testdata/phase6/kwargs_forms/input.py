def keys(**kw: int) -> list[str]:
    return sorted(kw)


def total(**kw: int) -> int:
    t: int = 0
    for k in kw:
        t += kw[k]
    return t


print(keys(), keys(b=1, a=2), total(x=1, y=2), total())
d: dict[str, int] = {"p": 1, "q": 2}
print(keys(**d), total(**d), total(**{"p": 1, "r": 3}))


def tag(name: str, **attrs: str) -> str:
    out: str = "<" + name
    for k, v in attrs.items():
        out += " " + k + '="' + v + '"'
    return out + ">"


print(tag("a"), tag("a", href="x", id="y"))

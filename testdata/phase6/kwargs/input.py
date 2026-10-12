def collect(**kw: int) -> int:
    t: int = 0
    for k in kw:
        t += kw[k]
    print(sorted(kw), len(kw))
    return t


print(collect(), collect(a=1), collect(a=1, b=2, c=3))


def settings(name: str, **options: str) -> str:
    parts: list[str] = [name]
    for key, value in options.items():
        parts.append(key + "=" + value)
    return ";".join(parts)


print(settings("x"), settings("x", color="red"), settings(name="y", mode="fast", level="3"))

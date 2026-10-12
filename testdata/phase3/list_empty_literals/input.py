def make() -> list[int]:
    return []


def total(xs: list[int]) -> int:
    t: int = 0
    for x in xs:
        t += x
    return t


xs: list[int] = make()
print(xs, total([]), total([1, 2]))
print(xs == [], [] == xs, xs != [], len([] + xs))
nested: list[list[int]] = [[], [1], []]
print(nested)
e: list[list[str]] = []
e.append([])
e[0].append("x")
print(e)
if xs:
    print("truthy")
else:
    print("falsy")
xs.append(1)
if xs:
    print("truthy")

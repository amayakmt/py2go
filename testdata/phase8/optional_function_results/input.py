from typing import Optional


def find(xs: list[int], target: int) -> Optional[int]:
    for i, v in enumerate(xs):
        if v == target:
            return i
    return None


def first_even(xs: list[int]) -> int | None:
    for v in xs:
        if v % 2 == 0:
            return v


def describe(n: Optional[int]) -> str:
    if n is None:
        return "nothing"
    return "got " + str(n + 1)


print(find([5, 6, 7], 7), find([5, 6, 7], 9))
print(first_even([1, 3, 4]), first_even([1, 3]))
print(describe(find([1], 1)), describe(find([1], 2)), describe(None), describe(7))
d: dict[str, int] = {"a": 1}
print(d.get("a"), d.get("b"), d.get("b", None), d.get("b", 0))
r: Optional[int] = d.get("a")
if r is not None:
    print(r + 10)

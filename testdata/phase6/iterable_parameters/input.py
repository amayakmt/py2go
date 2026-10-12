from typing import Iterable


def total(xs: Iterable[int]) -> int:
    t: int = 0
    for x in xs:
        t += x
    return t


def shout_all(words: Iterable[str]) -> list[str]:
    return [w.upper() for w in words]


print(total([1, 2, 3]), total({4, 5}), total(range(5)), total(x * 2 for x in range(3)))
print(total({1: "a", 2: "b"}), total([9]), total(map(len, ["ab", "c"])))
print(shout_all(["a", "b"]), shout_all("xy"), shout_all({"k": 1}))
d: dict[int, int] = {10: 1, 20: 2}
print(total(d), total(d.values()), total(d.keys()))
print(sorted(shout_all(w for w in ["b", "a"])))

from typing import Optional


def lookup(table: dict[str, int], key: str) -> Optional[int]:
    return table.get(key)


t: dict[str, int] = {"a": 1, "b": 0}
results: list[Optional[int]] = [lookup(t, k) for k in ["a", "b", "c"]]
print(results)
found: list[int] = [r for r in results if r is not None]
print(found, sum(found))
print([r is None for r in results], results.index(None), results.count(0))
pairs: dict[str, Optional[str]] = {"x": None, "y": "why"}
for k, v in pairs.items():
    if v is None:
        print(k, "has no value")
    else:
        print(k, v.upper())
best: Optional[int] = None
for n in [3, 9, 4]:
    if best is None or n > best:
        best = n
print(best)
name: Optional[str] = None
print("anonymous" if name is None else name)

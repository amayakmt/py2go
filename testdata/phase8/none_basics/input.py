from typing import Optional

x: Optional[int] = None
print(x, x is None, x is not None, x == None, bool(x))
x = 5
print(x, x is None, x is not None, x == 5, x != 6)
y: int | None = None
z: Optional[str] = "hi"
print(y, z, [y, 3], {"a": z})
print(None, [2])
names: list[Optional[str]] = ["a", None, "c"]
print(names, names.count(None), None in names)
print(f"{x} {y} {z!r}")
if x:
    print("truthy")
if not y:
    print("falsy")

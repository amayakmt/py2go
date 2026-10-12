from typing import Any, Union, Optional, Literal

def f(x: Any, y: Union[int, str], z: Optional[Union[int, str]] = None, w: Literal["a", "b"] = "a") -> Any:
    return [x, y, z, w]

print(f(1, "s"), f([1], 2, None, "b"))
v: int | str = 3
v = "now a string"
print(v)
items: list[Any] = [1, "a", None]
items.append(2.5)
print(items, f.__name__)

from typing import Optional


class Box:
    def __init__(self) -> None:
        self.item: Optional[list[int]] = None
        self.name: Optional[str] = None


b: Box = Box()
try:
    print(b.item[0])
except TypeError as e:
    print("TypeError:", e)
try:
    for x in b.item:
        print(x)
except TypeError as e:
    print("TypeError:", e)
try:
    print(b.name.upper())
except AttributeError as e:
    print("AttributeError:", e)
b.name = "ok"
print(b.name.upper(), b.name)

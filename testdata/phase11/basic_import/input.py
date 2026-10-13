import mathx
from mathx import area, Shape as S
import mathx as m2

print(mathx.PI, area(2), m2.area(1))
print(mathx.bump(), mathx.bump(), mathx.counter)
s = S("sq")
print(s.describe(), mathx.Shape("tri").describe(), __name__, mathx.__name__)
print(type(s).__name__, isinstance(s, mathx.Shape))

import a.b.c.leaf
from a.b import c
from a.b.c import leaf_value
print(a.b.c.leaf.value, c.leaf_value, leaf_value, a.b.c.leaf.where())
import a.b as ab
print(ab.leaf.where(), ab.c.leaf.value)

import bigs
from bigs import LIMIT, grow
print(LIMIT, grow(3), bigs.grow(LIMIT), bigs.hexed(LIMIT), LIMIT == bigs.LIMIT, bigs.LIMIT > 10 ** 20)

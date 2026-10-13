from util import add, sub as subtract
from util import *
from plain import *
import util
print(add(1, 2), subtract(5, 3), VERSION, Counter().inc(), util.hidden, util._private)
print(alpha, beta, f(), util.sub(9, 1))
try:
    print(_gamma)
except NameError as e:
    print("NameError", e)
c = Counter()
c.inc(); c.inc()
print(c.n, type(c).__name__, add("a", "b"), add([1], [2]), add(2 ** 70, 1))
beta.append(3)
import plain
print(plain.beta, plain.alpha, plain._gamma)

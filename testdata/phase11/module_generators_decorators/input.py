from deco import logged, squares, make_adder
import deco

@logged
def add(a, b):
    return a + b

@deco.logged
def neg(a):
    return -a

print(add(1, 2), neg(5))
print(list(squares(5)), sum(deco.squares(4)), make_adder(3)(4), deco.make_adder(1)(1))
print([x for x in squares(4) if x], {k: deco.make_adder(k)(1) for k in range(3)})

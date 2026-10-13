import operator as op
print(op.add(2, 3), op.sub(2, 3), op.mul(2, 3), op.truediv(7, 2), op.floordiv(7, 2), op.mod(7, 3), op.pow(2, 10), op.neg(5), op.pos(-5), op.abs(-7), op.inv(5), op.invert(0), op.lshift(1, 70), op.rshift(2 ** 70, 68), op.and_(12, 10), op.or_(12, 10), op.xor(12, 10))
print(op.lt(1, 2), op.le(2, 2), op.eq(1, 1.0), op.ne("a", "b"), op.ge(1, 2), op.gt(3, 2), op.not_(0), op.truth([]), op.truth("x"), op.is_(None, None), op.is_not(1, None), op.contains("abc", "b"), op.contains([1, 2], 3))
print(op.concat("ab", "cd"), op.concat([1], [2]), op.countOf([1, 2, 1, 1], 1), op.indexOf("hello", "l"), op.getitem([5, 6, 7], 1), op.getitem({"a": 1}, "a"), op.getitem("abc", slice(0, 2)))
xs = [1, 2, 3]
op.setitem(xs, 0, 10)
op.delitem(xs, 1)
d = {"k": 1}
op.setitem(d, "z", 2)
print(xs, d, op.iadd([1], [2]), op.isub(5, 2), op.imul(3, 4), op.iconcat("a", "b"), op.ior({1}, {2}))
people = [("bob", 25, "x"), ("alice", 30, "y"), ("carl", 25, "z")]
print(sorted(people, key=op.itemgetter(1)), sorted(people, key=op.itemgetter(1, 0)), op.itemgetter(0, 2)(people[0]), list(map(op.itemgetter(0), people)), op.itemgetter("a")({"a": 9}))
class P:
    def __init__(self, x):
        self.x = x
        self.inner = self
print(op.attrgetter("x")(P(3)), op.attrgetter("x", "inner.x")(P(4)), sorted([P(3), P(1)], key=op.attrgetter("x"))[0].x, op.methodcaller("upper")("abc"), op.methodcaller("replace", "a", "b")("banana"), op.methodcaller("split", ",")("a,b"))
for f, a in ((op.add, (1, "a")), (op.getitem, ([1], 5)), (op.truediv, (1, 0)), (op.lt, (1, "a")), (op.indexOf, ([1], 2)), (op.neg, ("a",))):
    try:
        print(f(*a))
    except (TypeError, IndexError, ZeroDivisionError, ValueError) as e:
        print(type(e).__name__, e)

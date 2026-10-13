import one, two
from one import helper as h1
from two import helper as h2
VERSION = "main"
def helper(x):
    return x - 1
count = 30
print(one.VERSION, two.VERSION, VERSION, h1(5), h2(5), helper(5), one.helper(1), two.helper(1))
print(one.Thing.kind, two.Thing.kind, one.count, two.count, count, type(one.Thing()).__name__)
def shadow():
    count = 99
    helper = lambda v: v
    return count, helper(7), one.count
print(shadow(), count)

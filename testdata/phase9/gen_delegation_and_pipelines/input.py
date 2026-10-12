def numbers():
    yield 1
    yield 2
    return "done"

def wrapper():
    result = yield from numbers()
    yield result
    yield from [10, 20]
    yield from "ab"

print(list(wrapper()))

def evens(it):
    for x in it:
        if x % 2 == 0:
            yield x

def squares(it):
    for x in it:
        yield x * x

pipeline = squares(evens(range(10)))
print(list(pipeline), list(pipeline))
print(list(squares(evens([1, 2, 3, 4]))), sum(squares(evens(range(6)))))
gen = (x for x in range(3))
print(list(gen), list(gen))
def flatten(items):
    for item in items:
        if isinstance(item, (list, tuple)):
            yield from flatten(item)
        else:
            yield item
print(list(flatten([1, [2, (3, [4, 5])], 6, [[7]]])))

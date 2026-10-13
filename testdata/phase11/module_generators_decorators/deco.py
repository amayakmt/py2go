import functools_lite
def logged(f):
    def wrapper(*args):
        print("call", f.__name__, args)
        return f(*args)
    return wrapper

def squares(n):
    for i in range(n):
        yield i * i

def make_adder(k):
    return lambda x: x + k

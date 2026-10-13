"""functools: reduce, partial, caches and the helpers around them."""


def reduce(function, iterable, *initial):
    it = iter(iterable)
    if len(initial) > 1:
        raise TypeError("reduce expected at most 3 arguments, got " + str(len(initial) + 2))
    if initial:
        value = initial[0]
    else:
        try:
            value = next(it)
        except StopIteration:
            raise TypeError("reduce() of empty iterable with no initial value") from None
    for element in it:
        value = function(value, element)
    return value


def partial(func, /, *args, **keywords):
    def call(*more, **more_kw):
        merged = dict(keywords)
        merged.update(more_kw)
        return func(*args, *more, **merged)
    return call


class _CmpKey:
    def __init__(self, mycmp, obj):
        self.mycmp = mycmp
        self.obj = obj

    def __lt__(self, other):
        return self.mycmp(self.obj, other.obj) < 0

    def __gt__(self, other):
        return self.mycmp(self.obj, other.obj) > 0

    def __eq__(self, other):
        return self.mycmp(self.obj, other.obj) == 0

    def __le__(self, other):
        return self.mycmp(self.obj, other.obj) <= 0

    def __ge__(self, other):
        return self.mycmp(self.obj, other.obj) >= 0


def cmp_to_key(mycmp):
    return lambda obj: _CmpKey(mycmp, obj)


def cache(user_function):
    store = {}

    def wrapper(*args):
        if args in store:
            return store[args]
        result = user_function(*args)
        store[args] = result
        return result
    return wrapper


def lru_cache(maxsize=128, typed=False):
    if callable(maxsize) and not isinstance(maxsize, int):
        return cache(maxsize)

    def decorating(user_function):
        return cache(user_function)
    return decorating


def wraps(wrapped, assigned=None, updated=None):
    def decorating(wrapper):
        return wrapper
    return decorating


def update_wrapper(wrapper, wrapped, assigned=None, updated=None):
    return wrapper

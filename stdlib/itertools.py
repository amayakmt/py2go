"""itertools, written with generators."""


def count(start=0, step=1):
    n = start
    while True:
        yield n
        n += step


def cycle(iterable):
    saved = []
    for x in iterable:
        yield x
        saved.append(x)
    while saved:
        for x in saved:
            yield x


def repeat(obj, times=None):
    if times is None:
        while True:
            yield obj
    else:
        for _ in range(times):
            yield obj


def chain(*iterables):
    for it in iterables:
        for x in it:
            yield x


def chain__from_iterable(iterables):
    for it in iterables:
        for x in it:
            yield x


def islice(iterable, *args):
    if len(args) == 1:
        start, stop, step = 0, args[0], 1
    elif len(args) == 2:
        start, stop, step = args[0], args[1], 1
    elif len(args) == 3:
        start, stop, step = args[0], args[1], args[2]
    else:
        raise TypeError("islice expected at most 4 arguments, got " + str(len(args) + 1))
    if start is None:
        start = 0
    if step is None:
        step = 1
    if len(args) == 1:
        if stop is not None and stop < 0:
            raise ValueError("Stop argument for islice() must be None or an integer: 0 <= x <= sys.maxsize.")
    elif start < 0 or (stop is not None and stop < 0):
        raise ValueError("Indices for islice() must be None or an integer: 0 <= x <= sys.maxsize.")
    if step <= 0:
        raise ValueError("Step for islice() must be a positive integer or None.")
    it = iter(iterable)
    i = 0
    next_i = start
    while stop is None or i < stop:
        try:
            x = next(it)
        except StopIteration:
            return
        if i == next_i:
            yield x
            next_i += step
        i += 1


def accumulate(iterable, func=None, *, initial=None):
    it = iter(iterable)
    total = initial
    if initial is None:
        try:
            total = next(it)
        except StopIteration:
            return
    yield total
    for x in it:
        if func is None:
            total = total + x
        else:
            total = func(total, x)
        yield total


def starmap(function, iterable):
    for args in iterable:
        yield function(*args)


def takewhile(predicate, iterable):
    for x in iterable:
        if predicate(x):
            yield x
        else:
            break


def dropwhile(predicate, iterable):
    dropping = True
    for x in iterable:
        if dropping and predicate(x):
            continue
        dropping = False
        yield x


def filterfalse(predicate, iterable):
    for x in iterable:
        if predicate is None:
            if not x:
                yield x
        elif not predicate(x):
            yield x


def compress(data, selectors):
    for d, s in zip(data, selectors):
        if s:
            yield d


def zip_longest(*iterables, fillvalue=None):
    iterators = [iter(it) for it in iterables]
    active = len(iterators)
    if not active:
        return
    while True:
        values = []
        for i in range(len(iterators)):
            it = iterators[i]
            if it is None:
                values.append(fillvalue)
                continue
            try:
                values.append(next(it))
            except StopIteration:
                active -= 1
                if not active:
                    return
                iterators[i] = None
                values.append(fillvalue)
        yield tuple(values)


def product(*iterables, repeat=1):
    pools = [tuple(p) for p in iterables] * repeat
    result = [[]]
    for pool in pools:
        result = [x + [y] for x in result for y in pool]
    for prod in result:
        yield tuple(prod)


def permutations(iterable, r=None):
    pool = tuple(iterable)
    n = len(pool)
    if r is None:
        r = n
    if r > n:
        return
    indices = list(range(n))
    cycles = list(range(n, n - r, -1))
    yield tuple(pool[i] for i in indices[:r])
    while n:
        for i in reversed(range(r)):
            cycles[i] -= 1
            if cycles[i] == 0:
                indices[i:] = indices[i + 1:] + indices[i:i + 1]
                cycles[i] = n - i
            else:
                j = cycles[i]
                indices[i], indices[-j] = indices[-j], indices[i]
                yield tuple(pool[i] for i in indices[:r])
                break
        else:
            return


def combinations(iterable, r):
    pool = tuple(iterable)
    n = len(pool)
    if r > n:
        return
    indices = list(range(r))
    yield tuple(pool[i] for i in indices)
    while True:
        for i in reversed(range(r)):
            if indices[i] != i + n - r:
                break
        else:
            return
        indices[i] += 1
        for j in range(i + 1, r):
            indices[j] = indices[j - 1] + 1
        yield tuple(pool[i] for i in indices)


def combinations_with_replacement(iterable, r):
    pool = tuple(iterable)
    n = len(pool)
    if not n and r:
        return
    indices = [0] * r
    yield tuple(pool[i] for i in indices)
    while True:
        for i in reversed(range(r)):
            if indices[i] != n - 1:
                break
        else:
            return
        indices[i:] = [indices[i] + 1] * (r - i)
        yield tuple(pool[i] for i in indices)


def groupby(iterable, key=None):
    it = iter(iterable)
    try:
        value = next(it)
    except StopIteration:
        return
    while True:
        k = value if key is None else key(value)
        group = [value]
        done = True
        for value in it:
            kk = value if key is None else key(value)
            if kk == k:
                group.append(value)
            else:
                done = False
                break
        yield (k, iter(group))
        if done:
            return


def pairwise(iterable):
    it = iter(iterable)
    try:
        a = next(it)
    except StopIteration:
        return
    for b in it:
        yield (a, b)
        a = b


def batched(iterable, n):
    if n < 1:
        raise ValueError("n must be at least one")
    it = iter(iterable)
    while True:
        batch = tuple(islice(it, n))
        if not batch:
            return
        yield batch


def tee(iterable, n=2):
    saved = list(iterable)
    return tuple(iter(saved) for _ in range(n))

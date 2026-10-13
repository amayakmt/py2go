"""heapq: the heap queue algorithm."""


def heappush(heap, item):
    heap.append(item)
    _siftdown(heap, 0, len(heap) - 1)


def heappop(heap):
    if not heap:
        raise IndexError("index out of range")
    lastelt = heap.pop()
    if heap:
        returnitem = heap[0]
        heap[0] = lastelt
        _siftup(heap, 0)
        return returnitem
    return lastelt


def heapreplace(heap, item):
    if not heap:
        raise IndexError("index out of range")
    returnitem = heap[0]
    heap[0] = item
    _siftup(heap, 0)
    return returnitem


def heappushpop(heap, item):
    if heap and heap[0] < item:
        item, heap[0] = heap[0], item
        _siftup(heap, 0)
    return item


def heapify(x):
    n = len(x)
    for i in reversed(range(n // 2)):
        _siftup(x, i)


def _siftdown(heap, startpos, pos):
    newitem = heap[pos]
    while pos > startpos:
        parentpos = (pos - 1) >> 1
        parent = heap[parentpos]
        if newitem < parent:
            heap[pos] = parent
            pos = parentpos
            continue
        break
    heap[pos] = newitem


def _siftup(heap, pos):
    endpos = len(heap)
    startpos = pos
    newitem = heap[pos]
    childpos = 2 * pos + 1
    while childpos < endpos:
        rightpos = childpos + 1
        if rightpos < endpos and not heap[childpos] < heap[rightpos]:
            childpos = rightpos
        heap[pos] = heap[childpos]
        pos = childpos
        childpos = 2 * pos + 1
    heap[pos] = newitem
    _siftdown(heap, startpos, pos)


def nlargest(n, iterable, key=None):
    if key is None:
        return sorted(iterable, reverse=True)[:n]
    return sorted(iterable, key=key, reverse=True)[:n]


def nsmallest(n, iterable, key=None):
    if key is None:
        return sorted(iterable)[:n]
    return sorted(iterable, key=key)[:n]


def merge(*iterables, key=None, reverse=False):
    items = []
    for it in iterables:
        for x in it:
            items.append(x)
    items.sort(key=key, reverse=reverse)
    for x in items:
        yield x

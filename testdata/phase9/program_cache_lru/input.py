class LRU:
    def __init__(self, capacity):
        self.capacity = capacity
        self.data = {}
        self.order = []
        self.hits = self.misses = 0

    def get(self, key, default=None):
        if key in self.data:
            self.order.remove(key)
            self.order.append(key)
            self.hits += 1
            return self.data[key]
        self.misses += 1
        return default

    def put(self, key, value):
        if key in self.data:
            self.order.remove(key)
        elif len(self.data) >= self.capacity:
            oldest = self.order.pop(0)
            del self.data[oldest]
        self.data[key] = value
        self.order.append(key)

c = LRU(2)
c.put("a", 1)
c.put("b", 2)
print(c.get("a"), c.get("z"), c.get("z", 0))
c.put("c", 3)
print(c.get("b"), c.get("a"), c.get("c"), c.order)
c.put("a", 10)
c.put("d", 4)
print(sorted(c.data.items()), c.order, c.hits, c.misses)

def memoize(f):
    cache = {}
    def wrapper(*args):
        if args not in cache:
            cache[args] = f(*args)
        return cache[args]
    return wrapper

@memoize
def slow_square(n):
    return n * n
print([slow_square(n) for n in (3, 4, 3, 5, 4)])

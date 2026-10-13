import io
import pickle


class Counter:
    def __init__(self, start):
        self.n = start
        self.log = []

    def bump(self):
        self.n += 1
        self.log.append(self.n)


class WithState:
    def __init__(self, a, b):
        self.a = a
        self.b = b
        self.cache = "big"

    def __getstate__(self):
        return {"a": self.a, "b": self.b}

    def __setstate__(self, st):
        self.a = st["a"]
        self.b = st["b"]
        self.cache = "rebuilt"


class Reducer:
    def __init__(self, v):
        self.v = v

    def __reduce__(self):
        return (Reducer, (self.v * 2,))


class NewArgs:
    def __init__(self, k):
        self.k = k

    def __getnewargs__(self):
        return (self.k,)


class MyError(Exception):
    def __init__(self, msg, code):
        super().__init__(msg, code)
        self.code = code


class Node:
    def __init__(self, name):
        self.name = name
        self.peer = None


def helper(x):
    return x + 1


c = Counter(5)
c.bump()
c.bump()
for proto in range(6):
    d = pickle.dumps(c, proto)
    print(proto, d)
    c2 = pickle.loads(d)
    print(c2.n, c2.log, type(c2).__name__)
    c2.bump()
    print(c2.n, c.n)

for proto in (0, 2, 4):
    w = pickle.loads(pickle.dumps(WithState(1, 2), proto))
    print(w.a, w.b, w.cache)
    r = pickle.loads(pickle.dumps(Reducer(21), proto))
    print(r.v)
    print(pickle.dumps(Reducer(1), proto))
    n = pickle.loads(pickle.dumps(NewArgs(7), proto))
    print(n.k)

a = Node("a")
b = Node("b")
a.peer = b
b.peer = a
for proto in (2, 4, 5):
    d = pickle.dumps(a, proto)
    print(d)
    a2 = pickle.loads(d)
    print(a2.name, a2.peer.name, a2.peer.peer is a2)

for proto in (0, 2, 4):
    d = pickle.dumps(helper, proto)
    print(d)
    f = pickle.loads(d)
    print(f(1), f is helper)
    d = pickle.dumps(MyError("boom", 42), proto)
    print(d)
    e = pickle.loads(d)
    print(type(e).__name__, e.args, e.code)

buf = io.BytesIO()
p = pickle.Pickler(buf, 4)
p.dump([1, 2])
p.dump({"k": (1, 2)})
p.dump("end")
buf.seek(0)
u = pickle.Unpickler(buf)
print(u.load(), u.load(), u.load())
try:
    u.load()
except EOFError as e:
    print("EOF", e)

buf = io.BytesIO()
pickle.dump({"x": [1, 2, 3]}, buf)
buf.seek(0)
print(pickle.load(buf))

class Unregistered:
    pass


data = pickle.dumps(Unregistered())
print(data)
print(type(pickle.loads(data)).__name__)
for bad in [b"cnosuchmod\nthing\n.", b"c__main__\nMissing\n.", b"\x80\x04\x95\x00\x00\x00\x00\x00\x00\x00\x00.", b"]a.", b"0.", b"\x80\x02K\x01"]:
    try:
        print(pickle.loads(bad))
    except Exception as e:
        print(type(e).__name__, str(e)[:40])
try:
    pickle.dumps(lambda x: x)
except Exception as e:
    print(type(e).__name__)
print(issubclass(pickle.PicklingError, pickle.PickleError), issubclass(pickle.UnpicklingError, pickle.PickleError))
print(pickle.loads(pickle.dumps(10**30)), pickle.loads(pickle.dumps(-10**30, 2)))

from contextlib import contextmanager, suppress, closing, nullcontext, ExitStack
import contextlib
events = []
@contextmanager
def tag(name):
    events.append("open " + name)
    try:
        yield name.upper()
    finally:
        events.append("close " + name)
with tag("a") as x, tag("b") as y:
    events.append(x + y)
print(events)
events.clear()
try:
    with tag("c"):
        raise KeyError("k")
except KeyError as e:
    events.append("caught " + str(e))
print(events)
@contextmanager
def swallow():
    try:
        yield
    except ValueError:
        events.append("swallowed")
events.clear()
with swallow():
    raise ValueError("x")
print(events)
with suppress(ZeroDivisionError):
    print(1 / 0)
with suppress(KeyError, ValueError):
    int("z")
print("after")
class Res:
    def __init__(self):
        self.closed = False
    def close(self):
        self.closed = True
r = Res()
with closing(r) as c:
    print(c is r, r.closed)
print(r.closed)
with nullcontext("enter") as v, nullcontext():
    print(v)
class Mgr:
    def __init__(self, name):
        self.name = name
    def __enter__(self):
        print("enter", self.name)
        return self.name
    def __exit__(self, t, v, tb):
        print("exit", self.name)
        return False
with ExitStack() as stack:
    a = stack.enter_context(Mgr("one"))
    b = stack.enter_context(Mgr("two"))
    stack.callback(print, "callback", a, b)
    print("body")
print("end")

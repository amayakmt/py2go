class Emitter:
    def __init__(self):
        self.handlers = {}

    def on(self, event, handler):
        self.handlers.setdefault(event, []).append(handler)
        return handler

    def emit(self, event, *args, **kwargs):
        results = []
        for h in self.handlers.get(event, []):
            results.append(h(*args, **kwargs))
        return results

em = Emitter()
log = []

@em.on_decorator if False else (lambda f: f)
def unused():
    pass

def on_greet(name, punct="!"):
    log.append("hello " + name + punct)
    return len(name)

em.on("greet", on_greet)
em.on("greet", lambda name, punct="": log.append("again " + name) or 0)
em.on("sum", lambda *nums: sum(nums))
em.on("sum", lambda *nums: max(nums))
em.on("kw", lambda **kw: sorted(kw.items()))

print(em.emit("greet", "ann"), em.emit("greet", "bo", punct="?"), em.emit("nothing"))
print(em.emit("sum", 1, 2, 3), em.emit("kw", b=2, a=1))
print(log)
counter = {"n": 0}
def count(*args):
    counter["n"] += 1
    return counter["n"]
for _ in range(3):
    em.on("tick", count)
print(em.emit("tick"), counter)

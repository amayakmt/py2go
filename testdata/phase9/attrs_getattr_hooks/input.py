class Config:
    def __init__(self, **values):
        self._values = values

    def __getattr__(self, name):
        if name in self._values:
            return self._values[name]
        raise AttributeError("no setting " + name)

cfg = Config(debug=True, level=3)
print(cfg.debug, cfg.level)
print(getattr(cfg, "level"), hasattr(cfg, "debug"), hasattr(cfg, "nope"), getattr(cfg, "nope", None))
try:
    cfg.missing
except AttributeError as e:
    print(e)

class Proxy:
    def __init__(self, target):
        self.target = target
        self.calls = []

    def __getattr__(self, name):
        attr = getattr(self.target, name)
        self.calls.append(name)
        return attr

p = Proxy([3, 1, 2])
p.append(0)
p.sort()
print(p.target, p.calls, p.count(1))

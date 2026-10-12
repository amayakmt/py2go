counter = 0
registry = {}
names = []
config = None
LIMIT = 3

def bump():
    global counter
    counter += 1
    return counter

def register(name, value):
    registry[name] = value
    names.append(name)

def configure(**kwargs):
    global config
    config = dict(kwargs)

print(bump(), bump(), counter)
register("a", 1)
register("b", [2])
print(registry, names)
print(config)
configure(debug=True, level=LIMIT)
print(config, config["level"] * 2)
for i in range(LIMIT):
    register("n%d" % i, i)
print(len(registry), names[-1], counter < LIMIT)
total = sum(v for v in registry.values() if isinstance(v, int))
print(total)

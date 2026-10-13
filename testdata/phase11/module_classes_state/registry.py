_registry = {}

def register(name, cls):
    _registry[name] = cls
    return cls

def create(name, *args):
    return _registry[name](*args)

def names():
    return sorted(_registry)

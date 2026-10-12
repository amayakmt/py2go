wrapper_calls = {}

def logged(fn):
    calls = []
    def wrapper(*args, **kwargs):
        calls.append((args, sorted(kwargs.items())))
        return fn(*args, **kwargs)
    wrapper_calls[fn.__name__] = calls
    return wrapper

def repeat(times):
    def decorate(fn):
        def wrapper(*args, **kwargs):
            return [fn(*args, **kwargs) for _ in range(times)]
        return wrapper
    return decorate

@logged
def add(a, b=0):
    return a + b

@repeat(3)
def hello(name):
    return "hi " + name

print(add(1, 2), add(5), add(a=1, b=9))
print(wrapper_calls["add"])
print(hello("bob"))

@logged
@repeat(2)
def twice(x):
    return x * 2
print(twice(4), wrapper_calls["wrapper"])

def register(registry):
    def decorate(fn):
        registry[fn.__name__] = fn
        return fn
    return decorate

handlers = {}
@register(handlers)
def on_start():
    return "started"
@register(handlers)
def on_stop():
    return "stopped"
print(sorted(handlers), handlers["on_start"](), on_stop())

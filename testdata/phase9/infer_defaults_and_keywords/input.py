def greet(name, greeting="Hello", punctuation="!"):
    return greeting + ", " + name + punctuation

print(greet("Ada"))
print(greet("Ada", "Hi"))
print(greet("Ada", punctuation="?"))
print(greet(greeting="Yo", name="Bob"))

def summary(*args, **kwargs):
    return len(args), sorted(kwargs.items())

print(summary())
print(summary(1, 2, 3, a=1, b=2))

def scaled(values, factor=2, offset=0):
    return [v * factor + offset for v in values]

print(scaled([1, 2, 3]), scaled([1, 2], 3, 1), scaled((1.5,), offset=0.5))
opts = {"factor": 10, "offset": 1}
print(scaled([1, 2], **opts))
nums = [4, 5]
print(summary(*nums, **{"z": 26}))

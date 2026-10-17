class Frozen:
    def __init__(self, x, y):
        object.__setattr__(self, "x", x)
        object.__setattr__(self, "y", y)

    def __setattr__(self, name, value):
        raise AttributeError("frozen: " + name)


f = Frozen(1, 2)
print(f.x, f.y)
try:
    f.x = 5
except AttributeError as e:
    print("AttributeError", e)
print(f.x)


class Validated:
    def __init__(self):
        self.log = []
        self.age = 0

    def __setattr__(self, name, value):
        if name == "age" and value < 0:
            raise ValueError("age must be >= 0")
        super().__setattr__(name, value)


v = Validated()
v.age = 30
print(v.age)
try:
    v.age = -1
except ValueError as e:
    print("ValueError", e)
print(v.age)


class Tracked:
    def __init__(self):
        self.changes = []
        self.value = 1

    def __setattr__(self, name, value):
        if name != "changes":
            self.changes.append(name)
        object.__setattr__(self, name, value)


t = Tracked()
t.value = 2
t.value = 3
setattr(t, "value", 4)
print(t.value, t.changes)

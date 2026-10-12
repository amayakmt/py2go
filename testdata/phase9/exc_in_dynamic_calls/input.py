class Resource:
    def __init__(self, name):
        self.name = name
        self.open = True

    def close(self):
        self.open = False

    def __enter__(self):
        return self

    def __exit__(self, kind, value, tb):
        self.close()
        return kind is not None and issubclass(kind, KeyError)

def run(action, *args):
    with Resource("r") as res:
        try:
            return action(res, *args)
        except ZeroDivisionError as e:
            return "zde: " + str(e)
        finally:
            print("closing", res.name)

print(run(lambda r, a, b: a / b, 6, 3))
print(run(lambda r, a, b: a / b, 1, 0))
print(run(lambda r, key: {}[key], "missing"))
try:
    print(run(lambda r: [][0]))
except IndexError as e:
    print("IndexError escapes:", e)
def chain():
    try:
        raise ValueError("inner")
    except ValueError as e:
        raise TypeError("outer") from e
try:
    chain()
except TypeError as e:
    print(e, repr(e.__cause__), type(e.__cause__).__name__)
errors = []
for f in (int, float, len, abs, str.upper):
    try:
        f(None)
    except (TypeError, AttributeError) as e:
        errors.append(type(e).__name__)
print(errors)

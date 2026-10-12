class AppError(Exception):
    def __init__(self, message, code=500):
        super().__init__(message)
        self.message = message
        self.code = code

class NotFound(AppError):
    def __init__(self, what):
        super().__init__(what + " not found", 404)
        self.what = what

def lookup(table, key):
    if key not in table:
        raise NotFound(key)
    return table[key]

table = {"a": 1}
for key in ("a", "b"):
    try:
        print(lookup(table, key))
    except NotFound as e:
        print(e.what, e.code, e.message, str(e), isinstance(e, AppError), e.args)
    except AppError:
        print("other")
    finally:
        print("checked", key)

errors = []
for exc in (AppError("x"), NotFound("y"), ValueError("z"), KeyError("k")):
    try:
        raise exc
    except (AppError, ValueError) as e:
        errors.append((type(e).__name__, str(e)))
    except Exception as e:
        errors.append(("other", repr(e)))
print(errors)
try:
    try:
        lookup({}, "q")
    except NotFound as e:
        raise RuntimeError("wrapped") from e
except RuntimeError as e:
    print(e, type(e.__cause__).__name__, e.__cause__.code)

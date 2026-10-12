errs: list[Exception] = [ValueError("a"), KeyError("b")]
for e in errs:
    try:
        raise e
    except ValueError as v:
        print("value error", v)
    except Exception as other:
        print("other", type(other).__name__, other)
e1: ValueError = ValueError("stored")
print(e1, repr(e1), e1.args)
pending: dict[str, Exception] = {"x": RuntimeError("boom")}
try:
    raise pending["x"]
except RuntimeError as r:
    print("caught", r)

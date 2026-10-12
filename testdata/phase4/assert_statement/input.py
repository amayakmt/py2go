def positive(n: int) -> int:
    assert n > 0, "n must be positive"
    return n


print(positive(3))
try:
    positive(-1)
except AssertionError as e:
    print("failed:", e, repr(e))
try:
    assert 1 + 1 == 3
except AssertionError as e:
    print("no message:", repr(str(e)))
try:
    assert False, 404
except AssertionError as e:
    print(repr(e))
xs: list[int] = []
try:
    assert xs, "list is empty"
except AssertionError as e:
    print(e)
assert [1], "never fails"
assert "x"
print("done")

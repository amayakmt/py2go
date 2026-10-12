base: dict[str, int] = {"a": 1, "b": 2}
extra: dict[str, int] = {"b": 20, "c": 30}
merged: dict[str, int] = {**base, **extra}
print(merged)
more: dict[str, int] = {"first": 0, **base, "b": 99}
print(more)

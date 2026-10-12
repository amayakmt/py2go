names: list[str] = []
try:
    raise ZeroDivisionError("a")
except Exception as e:
    names.append(type(e).__name__)
try:
    raise IndexError("a")
except Exception as e:
    names.append(type(e).__name__)
try:
    raise NotImplementedError("a")
except Exception as e:
    names.append(type(e).__name__)
try:
    raise AssertionError
except Exception as e:
    names.append(type(e).__name__)
try:
    print(1 // 0)
except Exception as e:
    names.append(type(e).__name__)
print(names)

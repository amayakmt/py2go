try:
    try:
        raise ValueError("first")
    finally:
        raise TypeError("second")
except Exception as e:
    print(repr(e))

try:
    raise ValueError("first")
except ValueError:
    print("handling")
    raise TypeError("while handling")

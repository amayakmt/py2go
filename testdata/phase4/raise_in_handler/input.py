try:
    try:
        raise ValueError("original")
    except ValueError as e:
        raise TypeError("replacement: " + str(e))
except TypeError as e:
    print(e)

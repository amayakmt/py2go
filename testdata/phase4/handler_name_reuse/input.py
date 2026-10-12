for k in range(2):
    try:
        raise ValueError(k)
    except ValueError as e:
        print(e)
try:
    raise TypeError("t")
except TypeError as e:
    print(e)
e: int = 5
print(e)

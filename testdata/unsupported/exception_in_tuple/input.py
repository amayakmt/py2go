try:
    pass
except ValueError as e:
    t: tuple[int, int] = (1, e)

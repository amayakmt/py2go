def early() -> int:
    return LATE


try:
    early()
except NameError as e:
    print(type(e).__name__, e)
LATE: int = 5
print(early())

def swallow() -> int:
    try:
        raise ValueError("lost")
    finally:
        return 7


def override() -> int:
    try:
        return 1
    finally:
        return 2


def stays() -> int:
    try:
        raise ValueError("kept")
    finally:
        print("finally runs")
    return 0


print(swallow())
print(override())
try:
    stays()
except ValueError as e:
    print("got", e)

def f() -> int:
    for i in range(5):
        try:
            if i == 2:
                return i * 100
        finally:
            print("cleanup", i)
    return -1


print(f())

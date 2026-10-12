def f() -> int:
    x: int = 1
    try:
        x = 2
        return x
    finally:
        x = 3
        print("finally sees", x)


def g() -> str:
    try:
        return "from try"
    finally:
        print("g cleanup")


print(f())
print(g())

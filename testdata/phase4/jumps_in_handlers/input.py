for i in range(5):
    try:
        if i % 2 == 1:
            raise ValueError(i)
        print("fine", i)
    except ValueError as e:
        if i == 3:
            print("stopping at", e)
            break
        print("skipping", e)
        continue
    print("after try", i)

def f(n: int) -> str:
    try:
        raise ValueError("x")
    except ValueError:
        return "returned from handler " + str(n)
    return "unreachable"

print(f(1))

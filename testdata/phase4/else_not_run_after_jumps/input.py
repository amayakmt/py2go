for i in range(3):
    try:
        if i == 0:
            continue
        if i == 2:
            break
        print("body", i)
    except ValueError:
        print("never")
    else:
        print("else", i)
    print("end of iteration", i)

def f(n: int) -> int:
    try:
        if n > 0:
            return n
    except ValueError:
        pass
    else:
        print("else runs only without return")
    return 0

print(f(1), f(-1))

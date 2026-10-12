def risky(n: int) -> int:
    if n == 1:
        raise ValueError("v")
    if n == 2:
        return 10 // (n - 2)
    return n


for i in range(4):
    try:
        print(risky(i))
    except:
        print("something failed at", i)

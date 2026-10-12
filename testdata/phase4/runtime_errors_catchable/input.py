def run(step: int) -> None:
    zero: int = 0
    fz: float = 0.0
    xs: list[int] = [1]
    empty: list[int] = []
    s: str = "abc"
    if step == 0:
        print(1 // zero)
    elif step == 1:
        print(1 % zero)
    elif step == 2:
        print(1 / zero)
    elif step == 3:
        print(1.5 / fz)
    elif step == 4:
        print(xs[5])
    elif step == 5:
        print(s[9])
    elif step == 6:
        print(int("x1"))
    elif step == 7:
        print(float("y"))
    elif step == 8:
        print(empty.pop())
    elif step == 9:
        xs.remove(7)
    elif step == 10:
        print(s.split(""))
    elif step == 11:
        print(max(empty))
    elif step == 12:
        print(2.0 ** 5000)
    elif step == 13:
        a, b = xs
    elif step == 14:
        print(xs.index(9))
    elif step == 15:
        print(1 << -1)
    elif step == 16:
        print(divmod(1, zero))
    elif step == 17:
        print(int(float("nan")))
    elif step == 18:
        print(0.0 ** -1)
    elif step == 19:
        print(xs[::zero])
    else:
        print("no error")


for i in range(21):
    try:
        run(i)
    except Exception as e:
        print(i, type(e).__name__, "-", e)

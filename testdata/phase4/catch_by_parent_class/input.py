def attempt(n: int) -> str:
    try:
        if n == 0:
            return str(1 // n)
        if n == 1:
            xs: list[int] = []
            return str(xs[n])
        if n == 2:
            return str(int("x"))
        if n == 3:
            return str(2.0 ** 10000)
        return "fine"
    except LookupError as e:
        return "lookup: " + str(e)
    except ArithmeticError as e:
        return "arithmetic: " + str(e)
    except Exception as e:
        return "other: " + str(e)


for i in range(5):
    print(i, attempt(i))

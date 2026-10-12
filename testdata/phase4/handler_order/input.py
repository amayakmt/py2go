def run(kind: int) -> None:
    try:
        if kind == 0:
            raise ValueError("v")
        if kind == 1:
            raise TypeError("t")
        if kind == 2:
            raise KeyError("k")
        if kind == 3:
            raise ZeroDivisionError("z")
        raise RuntimeError("r")
    except ValueError:
        print(kind, "ValueError handler")
    except (TypeError, KeyError) as e:
        print(kind, "TypeError or KeyError handler:", type(e).__name__)
    except ArithmeticError as e:
        print(kind, "ArithmeticError handler:", e)
    except Exception as e:
        print(kind, "Exception handler:", e)


for k in range(5):
    run(k)

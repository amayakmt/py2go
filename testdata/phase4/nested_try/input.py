def run(n: int) -> None:
    try:
        print("outer try")
        try:
            print("inner try")
            if n == 0:
                raise ValueError("inner problem")
            if n == 1:
                raise TypeError("escapes")
        except ValueError as e:
            print("inner caught", e)
            raise
        finally:
            print("inner finally")
    except ValueError as e:
        print("outer caught", e)
    except TypeError as e:
        print("outer caught other", e)
    finally:
        print("outer finally")


run(0)
run(1)
run(2)

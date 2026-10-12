def run(fail: bool) -> None:
    print("start", fail)
    try:
        print("try")
        if fail:
            raise ValueError("x")
        print("try end")
    except ValueError:
        print("except")
    else:
        print("else")
    finally:
        print("finally")
    print("after")


run(False)
run(True)

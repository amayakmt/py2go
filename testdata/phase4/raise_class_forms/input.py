for k in range(4):
    try:
        if k == 0:
            raise ZeroDivisionError
        if k == 1:
            raise ZeroDivisionError()
        if k == 2:
            raise OSError("disk")
        raise EOFError("end", 1)
    except Exception as e:
        print(k, repr(e), "|", e)

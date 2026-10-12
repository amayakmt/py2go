try:
    try:
        print("body")
    except ValueError:
        print("handler must not run")
    else:
        raise ValueError("raised in else")
except ValueError as e:
    print("outer got", e)

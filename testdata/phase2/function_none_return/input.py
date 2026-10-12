def say(msg: str) -> None:
    print(msg)


def early(n: int) -> None:
    if n < 0:
        print("negative")
        return
    if n == 0:
        print("zero")
        return None
    print("positive")


say("hi")
early(-1)
early(0)
early(1)

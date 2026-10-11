def a() -> None:
    print("from a")


def b() -> None:
    print("from b")
    a()


b()
a()
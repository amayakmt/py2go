def yes(label: str) -> bool:
    print("yes", label)
    return True


def no(label: str) -> bool:
    print("no", label)
    return False


print(yes("a") and yes("b"))
print(no("a") and yes("b"))
print(yes("a") or no("b"))
print(no("a") or no("b"))
print(yes("1") and yes("2") and no("3") and yes("4"))
print(no("1") or no("2") or yes("3") or yes("4"))
print(yes("x") and no("y") or yes("z"))
print(yes("x") or no("y") and no("z"))
print((yes("p") or no("q")) and no("r"))
print(not (yes("n") and no("m")))

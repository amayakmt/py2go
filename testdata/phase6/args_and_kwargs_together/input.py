def everything(a: int, b: int = 2, *args: int, flag: bool = False, **kw: int) -> str:
    return str(a) + str(b) + str(args) + str(flag) + str(kw)


print(everything(1))
print(everything(1, 3))
print(everything(1, 3, 5, 7))
print(everything(1, flag=True))
print(everything(1, 3, 5, flag=True, x=1, y=2))
print(everything(a=1, b=9, z=0))

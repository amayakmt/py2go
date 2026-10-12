def longer(a: str, b: str) -> str:
    if a > b:
        return a
    return b


print(longer("apple", "banana"))
print(longer("zebra", "ant"))
print(longer("", "x"))

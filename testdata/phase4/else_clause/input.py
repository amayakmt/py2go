def parse(s: str) -> int:
    try:
        n: int = int(s)
    except ValueError:
        print("  invalid")
        return -1
    else:
        print("  parsed")
        return n * 2
    finally:
        print("  finally")


print(parse("21"))
print(parse("x"))

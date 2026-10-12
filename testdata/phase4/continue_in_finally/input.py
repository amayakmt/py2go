for i in range(4):
    try:
        if i % 2 == 0:
            raise ValueError(i)
        print("odd", i)
    finally:
        print("fin", i)
        continue
print("done")

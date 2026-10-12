for i in range(5):
    try:
        if i == 1:
            continue
        if i == 3:
            break
        print("body", i)
    finally:
        print("finally", i)
print("loop done")

n: int = 0
while True:
    n += 1
    try:
        if n < 3:
            continue
        break
    finally:
        print("cleanup", n)
print(n)

for i in range(3):
    for j in range(3):
        if j == 2:
            break
        print(i, j)
    print("after inner", i)
i = 0
while i < 3:
    i += 1
    for j in range(5):
        if j > i:
            break
    print(i, j)

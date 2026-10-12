for i in range(4):
    if i == 1:
        continue
    for j in range(4):
        if j == 2:
            continue
        if j == 3:
            break
        print(i, j)

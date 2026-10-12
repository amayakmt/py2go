for q0 in range(1):
    if q0 == 3:
        break
else:
    print('no break', 1)
w1: int = 0
if (w1 := 2) > (3 + 5):
    print('big', w1)
else:
    print('small', w1)
w2: int = 0
if (w2 := ((1 - 3) * (3 + 8))) > (2 - (8 - 6)):
    print('big', w2)
else:
    print('small', w2)
for q3 in range(4):
    if q3 == 3:
        break
else:
    print('no break', 8)
print((9 + (8 + 1)) < ((8 + 6) + (9 * 2)) < 5, (9 + (8 + 1)) <= ((8 + 6) + (9 * 2)) != 5)
w5: int = 0
if (w5 := ((1 * 4) + (8 + 0))) > ((1 - 3) + 5):
    print('big', w5)
else:
    print('small', w5)
print(((8 * 9) + 6) if 4 == (6 - (5 * 0)) else 1)

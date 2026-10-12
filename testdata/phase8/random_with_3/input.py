w0: int = 0
if (w0 := 7) > 6:
    print('big', w0)
else:
    print('small', w0)
print((9 * (2 - 2)) if 5 <= (9 - 5) else ((4 * 3) + 4))
print(((4 * 1) - (1 * 4)) < ((9 + 5) + 4) < 1, ((4 * 1) - (1 * 4)) <= ((9 + 5) + 4) != 1)
print(((1 - 0) * 8) < 6 < ((8 - 5) - (8 * 8)), ((1 - 0) * 8) <= 6 != ((8 - 5) - (8 * 8)))
print(6 if ((0 + 1) + 9) == (9 - 8) else 8)
print(((0 * 8) + (1 - 5)) if (1 + 1) > 6 else (0 - (9 - 5)))
print(((5 - 1) * (6 * 7)) if ((3 + 9) * (2 - 8)) <= ((4 - 7) * (1 * 9)) else 8)
for q7 in range(4):
    if q7 == 4:
        break
else:
    print('no break', 1)
for q8 in range(4):
    if q8 == 3:
        break
else:
    print('no break', ((3 + 8) - 2))
for q9 in range(4):
    if q9 == 2:
        break
else:
    print('no break', 6)

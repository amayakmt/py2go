print(((9 * 3) * 3) if 6 < ((0 * 7) + 6) else 2)
w1: int = 0
if (w1 := 6) > (3 + (5 + 9)):
    print('big', w1)
else:
    print('small', w1)
for q2 in range(4):
    if q2 == 3:
        break
else:
    print('no break', 0)
print(6 < 2 < ((3 + 6) - (2 - 0)), 6 <= 2 != ((3 + 6) - (2 - 0)))
for q4 in range(3):
    if q4 == 3:
        break
else:
    print('no break', ((7 + 3) - 0))
w5: int = 0
if (w5 := 5) > ((4 - 0) - (7 + 1)):
    print('big', w5)
else:
    print('small', w5)
print(9 if ((9 - 3) - (0 - 9)) != ((3 - 6) - (7 - 5)) else 0)
for q7 in range(0):
    if q7 == 3:
        break
else:
    print('no break', ((6 * 4) * 2))
w8: int = 0
if (w8 := 5) > ((5 + 0) + 6):
    print('big', w8)
else:
    print('small', w8)
print(((3 * 1) - (6 - 9)) < ((0 * 5) - (6 + 8)) < 9, ((3 * 1) - (6 - 9)) <= ((0 * 5) - (6 + 8)) != 9)
print(7 < 4 < 3, 7 <= 4 != 3)

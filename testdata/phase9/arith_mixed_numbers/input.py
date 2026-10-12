def calc(a, b):
    return [a + b, a - b, a * b, a / b, a // b, a % b, a ** 2]

print(calc(7, 2))
print(calc(7.5, 2))
print(calc(-7, 2))
print(calc(7, -2.5))
print(calc(True, 2))
print(1 + True, True + True, 3 * False, 2 ** 10, 2 ** -1)
print(7 // -2, -7 // 2, 7 % -3, -7 % 3, 7.5 // 2, -7.5 % 2)
print(divmod(17, 5), divmod(-17, 5), round(2.5), round(3.5), round(-0.5), round(1.2345, 2))
print(abs(-3), abs(-3.5), abs(True), int(3.9), int(-3.9), float(2), int("12"), float("1.5"))
print(0.1 + 0.2, 1e3, 1 / 3, 10 / 2, 2 ** 0.5)

a = 10 ** 25 + 12345
b = 7 ** 20
print(a + b, a - b, b - a, a * b, a // b, a % b, divmod(a, b), divmod(-a, b), divmod(a, -b), divmod(-a, -b))
print(a // -7, a % -7, -a // 7, -a % 7, b // 0.5 if False else b / 3)
print(a & b, a | b, a ^ b, ~a, a << 3, a >> 3, -a >> 3, (a << 70) >> 70 == a)
print(abs(-a), -(-a), +a, a - a, a * 0, a * -1, a // a, a % a)
print(a / b, b / a, a / 3, a / 10 ** 20, 10 ** 400 / 10 ** 399)
print(float(a), float(b), int(1e22), int(-1e22), int(float(2 ** 80)), 2 ** 80 == float(2 ** 80))
print(a < 1e30, a > 1e20, a == 1e25, a + 0.5, a * 0.5, 1.5 * b)
try:
    print(a / 0)
except ZeroDivisionError as e:
    print("ZeroDivisionError", e)
try:
    print(a // 0)
except ZeroDivisionError as e:
    print("ZeroDivisionError", e)
try:
    print(a % 0)
except ZeroDivisionError as e:
    print("ZeroDivisionError", e)
try:
    print(float(10 ** 400))
except OverflowError as e:
    print("OverflowError", e)
try:
    print(1 << -1)
except ValueError as e:
    print("ValueError", e)

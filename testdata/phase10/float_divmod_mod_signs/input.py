vals = [7.5, -7.5, 0.5, -0.5, 1e20, 3.0, -3.0, 0.1]
divs = [2.0, -2.0, 0.3, -0.3, 1.5, 1e-5, 3, -3]
for a in vals:
    for b in divs:
        print(a, b, a // b, a % b, divmod(a, b), a / b)
for a, b in ((1.0, 0.0), (1, 0.0), (1.0, 0), (1.0, -0.0)):
    for f in (lambda: a / b, lambda: a // b, lambda: a % b, lambda: divmod(a, b)):
        try:
            print(f())
        except ZeroDivisionError as e:
            print("ZeroDivisionError", e)
print(-0.0 % 5, 0.0 % -5, 5 % -0.0 if False else 0, float("inf") % 5 if False else 0, 5 % float("inf"), -5 % float("inf"), 5.5 // float("inf"), -5.5 // float("inf"))

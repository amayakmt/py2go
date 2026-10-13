from fractions import Fraction
import math
a = Fraction(1, 3)
b = Fraction("2/5")
print(a, b, repr(a), a + b, a - b, a * b, a / b, a ** 2, a ** -2, 2 ** Fraction(2))
print(Fraction(0.75), Fraction("1.25e2"), Fraction(" -3/6 "), Fraction(3, -6), Fraction(7))
print(a + 1, 1 + a, 2 * a, 1 / a, a - 1, 1 - a, a + 0.5, 0.5 * a)
print(a < b, a == Fraction(2, 6), a != b, a > 0, a <= 1, Fraction(1, 2) == 0.5, 1 == Fraction(2, 2))
print(math.floor(Fraction(7, 2)), math.ceil(Fraction(7, 2)), round(Fraction(7, 2)), round(Fraction(5, 2)), int(Fraction(-7, 2)), math.trunc(Fraction(-7, 2)))
print(round(Fraction(12345, 100), 1), round(Fraction(12345, 100), -1), float(Fraction(1, 3)))
print(Fraction(3, 4).limit_denominator(2), Fraction(math.pi).limit_denominator(100), Fraction(1, 3).as_integer_ratio())
print(divmod(Fraction(7, 2), Fraction(1, 3)), Fraction(7, 2) // 1, Fraction(7, 2) % 1, -Fraction(1, 2), abs(Fraction(-1, 2)))
print(hash(Fraction(1, 2)) == hash(0.5), hash(Fraction(3)) == hash(3), {Fraction(1, 2): "x"}[Fraction(2, 4)])
print(f"{Fraction(1, 3):.4f}", f"{Fraction(22, 7):10.3f}|", f"{Fraction(1, 8):.1%}", format(Fraction(5, 3)), f"{Fraction(1,3):>8}")
for bad in ["abc", "1/0", ""]:
    try:
        Fraction(bad)
    except (ValueError, ZeroDivisionError) as e:
        print(type(e).__name__, e)
try:
    Fraction(1, 0)
except ZeroDivisionError as e:
    print(e)
print(sum([Fraction(1, n) for n in range(1, 8)]), Fraction(1, 3).is_integer(), Fraction(4, 2).is_integer())
print(max(Fraction(1, 3), Fraction(1, 2)), sorted([Fraction(3, 4), Fraction(1, 4), Fraction(1, 2)]))

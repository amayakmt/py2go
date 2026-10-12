def grade(score: int) -> str:
    if score >= 90:
        return "A"
    elif score >= 80:
        return "B"
    elif score >= 70:
        return "C"
    else:
        return "F"


print(grade(95))
print(grade(85))
print(grade(75))
print(grade(10))
n: int = 2
if n == 1:
    print("one")
elif n == 2:
    print("two")
elif n == 3:
    print("three")

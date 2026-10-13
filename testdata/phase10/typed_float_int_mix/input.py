def mean(xs: list[float]) -> float:
    return sum(xs) / len(xs)

def isqrt_float(n: int) -> int:
    return int(n ** 0.5)

def hyp(a: float, b: float) -> float:
    return (a * a + b * b) ** 0.5

nums: list[int] = [3, 1, 4, 1, 5, 9, 2, 6]
print(mean([float(n) for n in nums]), mean([1.5, 2.5]), isqrt_float(99), isqrt_float(10 ** 12), hyp(3.0, 4.0), hyp(1.0, 1.0))
a: float = 7.0
b: int = 2
print(a / b, a // b, a % b, divmod(a, b), a ** b, b ** a, round(a / 3, 3), int(a) // b, float(b), -a, abs(-a), 7.0 == 7, 7.0 is 7 if False else True)
print(10 / 4, 10 // 4, -10 // 4, 10 % -4, 2 ** -2, 2 ** 0.5, 0.1 + 0.2, 1e3, 1e-3, 5 / 2 * 2, 3 * 0.1, round(2.5), round(3.5), round(-0.5), round(1234.5678, -2), round(1234.5678, 2))
total: float = 0.0
for i in range(10):
    total += 0.1
print(total, total == 1.0, abs(total - 1.0) < 1e-9)

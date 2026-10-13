xs = [0.1] * 10
print(sum(xs), sum(xs) == 1.0, sum(xs, 0.0), sum([0.1, 0.2, 0.3]), sum([1e100, 1.0, -1e100]), sum([1, 2.5, 3]), sum([True, 2.0]))
acc = 0.0
for i in range(1, 101):
    acc += 1 / i
print(acc, repr(acc))
prod = 1.0
for i in range(1, 30):
    prod *= i
print(prod, int(prod), prod == 8841761993739701954543616000000)
print(sum(x * x for x in [0.5, 1.5, 2.5]), min([2.5, 1, 3]), max([2, 2.0]), max([2.0, 2]), sorted([3, 1.5, 2, 2.0, -1]), 1 + 1e-16 == 1, 1 + 2e-16 == 1)
mean = sum([1.5, 2.5, 4.0]) / 3
var = sum((x - mean) ** 2 for x in [1.5, 2.5, 4.0]) / 3
print(mean, var, var ** 0.5, 2 ** 0.5, 4 ** 0.5, (-8) ** 2.0, 10 ** -2, 10.0 ** 20, 2 ** -1074.0 if False else 0)

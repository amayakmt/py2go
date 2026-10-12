names: list[str] = ["a", "b", "c"]
nums: list[int] = [1, 2, 3]
for i, n in enumerate(names, 1):
    print(i, n)
print(list(zip(names, nums)), list(enumerate(names)))
print(dict(zip(names, nums)), dict(enumerate(names, 5)))
for a, b, c in zip(names, nums, [True, False, True]):
    print(a, b, c)
print(list(zip(names, nums[:2])), list(zip(names)))
print(sorted(zip(nums, names), reverse=True))
print([a * b for a, b in zip(nums, nums)])

words = ["banana", "Apple", "cherry", "apple", "Banana"]
print(sorted(words))
print(sorted(words, key=str.lower))
print(sorted(words, key=len, reverse=True))
print(sorted(words, key=lambda w: (len(w), w)))
nums = [3, -1, 2.5, -7, 0]
print(sorted(nums), sorted(nums, key=abs), sorted(nums, reverse=True))
pairs = [("b", 2), ("a", 2), ("c", 1)]
print(sorted(pairs, key=lambda p: (p[1], p[0])))
print(min(words), max(words), min(words, key=len), max(nums, key=abs))
words.sort(key=len)
print(words)
nums.sort(reverse=True)
print(nums, list(reversed(nums)))
print(sorted({"b": 1, "a": 2}), sorted({3, 1, 2}), sorted("hello"), sorted((3, 1, 2)))
print(sum([0.1] * 10), sum(range(5)), sum([[1], [2]], []), sum([1, 2, 3], 10))

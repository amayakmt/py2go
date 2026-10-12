xs: list[int] = [3, 1, 2, 5, 4]
ys: list[int] = sorted(xs)
print(xs, ys)
xs.sort()
print(xs)
xs.sort(reverse=True)
print(xs)
print(sorted(xs, reverse=True), sorted([2.5, -1.0, 0.0]))
words: list[str] = ["pear", "fig", "Apple", "apple", "banana"]
print(sorted(words))
words.sort(reverse=True)
print(words)
pairs: list[tuple[int, str]] = [(2, "b"), (1, "z"), (2, "a"), (1, "a")]
pairs.sort()
print(pairs)
pairs.sort(reverse=True)
print(pairs)
xs.reverse()
print(xs)
nested: list[list[int]] = [[2, 1], [1, 5], [1, 2, 3], [1], []]
nested.sort()
print(nested)
print(sorted([True, False, True]))

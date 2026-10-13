import itertools
from itertools import product, permutations, combinations, combinations_with_replacement
print(list(product("ab", [1, 2])), list(product([0, 1], repeat=3)), list(product()), list(product([])), len(list(product(range(3), repeat=4))), list(product("a", "b", "c")))
print(list(permutations(range(3))), list(permutations("abc", 2)), list(permutations([1, 2], 3)), list(permutations([], 0)), len(list(permutations(range(6)))), list(permutations(range(3), 0)))
print(list(combinations("abcd", 2)), list(combinations(range(4), 3)), list(combinations([1], 2)), list(combinations(range(3), 0)), len(list(combinations(range(10), 4))))
print(list(combinations_with_replacement("abc", 2)), list(combinations_with_replacement(range(2), 3)), len(list(combinations_with_replacement(range(5), 3))))
print(sorted(set(permutations("aab"))), [sum(c) for c in combinations([1, 2, 3, 4], 2)], max(product(range(3), range(3)), key=lambda p: p[0] * p[1]))
print(sum(1 for p in permutations(range(5)) if all(p[i] != i for i in range(5))), [c for c in combinations("abcde", 3) if "a" in c and "e" in c])

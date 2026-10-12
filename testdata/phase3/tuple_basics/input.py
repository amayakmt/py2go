t: tuple[int, str, float] = (1, "two", 3.0)
print(t, len(t), t[0], t[1], t[2], t[-1], t[-3])
one: tuple[int] = (5,)
print(one, len(one))
print((1, 2) == (1, 2), (1, 2) != (1, 3), (1, 2) < (1, 3), (1, 2) < (1, 2), (2, 0) > (1, 9), (1, "a") <= (1, "b"))
nested: tuple[tuple[int, int], str] = ((1, 2), "x")
print(nested, nested[0][1], nested[0])
pairs: list[tuple[int, int]] = [(3, 4), (1, 2)]
print(pairs, pairs[0], pairs[1][0])
flags: tuple[bool, bool] = (True, False)
print(flags)
if t:
    print("nonempty tuples are truthy")

import bisect
a = [1, 3, 3, 3, 5, 8]
print([bisect.bisect_left(a, x) for x in range(0, 10)], [bisect.bisect_right(a, x) for x in range(0, 10)], bisect.bisect(a, 3))
bisect.insort(a, 4); bisect.insort_left(a, 3); bisect.insort_right(a, 9)
print(a)
print(bisect.bisect_left(a, 3, 2), bisect.bisect_right(a, 3, 0, 3), bisect.bisect_left(["apple", "cherry"], "banana"), bisect.bisect_left([(1, "a"), (3, "c")], (2, "")))
people = [("ann", 20), ("bob", 30), ("cy", 40)]
print(bisect.bisect_left(people, 30, key=lambda p: p[1]), bisect.bisect_right(people, 30, key=lambda p: p[1]))
grades = "FDCBA"
def grade(score, cuts=[60, 70, 80, 90]):
    return grades[bisect.bisect(cuts, score)]
print([grade(s) for s in (33, 60, 77, 89, 90, 100)])
try:
    bisect.bisect_left([1], 1, -1)
except ValueError as e:
    print(e)

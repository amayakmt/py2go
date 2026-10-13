vals = [True, False, 1, 0, -1, 1.0, 0.0, 2 ** 70, 1j, 0j, None, "", "a", [], [0], (), (0,), {}, {0: 0}, set(), {0}, b"", b"a"]
for v in vals:
    print(repr(v), bool(v), not v, v and "yes", v or "no", (v and 1) or 2, [x for x in (v,) if x])
nums = [True, False, 1, 0, 2, -3, 1.5, 2 ** 65]
print([a + b for a in nums for b in nums][:20])
print([a * b for a in nums[:4] for b in nums[:4]], [a == b for a in nums[:4] for b in nums[:4]], [a is b for a in nums[:2] for b in nums[:2]])
print(sum(nums[:6]), sum(nums[:6], 1.5), max(nums[:6]), min(nums[:6]), sorted(nums[:6]), [type(a + b).__name__ for a in nums[:3] for b in nums[:3]], [type(a * 1.0).__name__ for a in nums[:3]])
print(True + True + True, True * 10, False + 0.5, -True, ~True, True // True, True % 3, True ** 3, divmod(True, 2), abs(True), True << 3, True | 2, True ^ 3, int(True) is 1 if False else 1, [True, False].count(1), [0, 1, True].index(True), {True: "t", 1: "u", 1.0: "v"})

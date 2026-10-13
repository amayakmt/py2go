t = str.maketrans("abc", "xyz")
print(t, str.maketrans({"a": "1", "b": None, "é": "ee", 99: 100}), str.maketrans("ab", "cd", "e"), "aabbccee".translate(t), "abc".translate({97: None}), "abc".translate({98: "BB"}), "abc".translate({}), "héllo".translate({ord("é"): "e", ord("l"): None}), "abc".translate([None, None] * 60 if False else {ord("c"): 0x1F600}))
for f in (lambda: str.maketrans("ab", "c"), lambda: str.maketrans("a"), lambda: str.maketrans({"ab": "x"}), lambda: str.maketrans(1, 2), lambda: str.maketrans("a", "b", 1), lambda: "a".translate({97: -1}), lambda: "a".translate({97: 0x110000}), lambda: "a".translate(5)):
    try:
        print(f())
    except (ValueError, TypeError, LookupError, OverflowError) as e:
        print(type(e).__name__, e)
vowels = str.maketrans("", "", "aeiou")
print("the quick brown fox".translate(vowels), "Hello, World!".translate(str.maketrans("lo", "01")), "".join(sorted(set("mississippi".translate(str.maketrans("s", "z"))))), str.maketrans("", "") == {}, len(str.maketrans("abc", "def")))

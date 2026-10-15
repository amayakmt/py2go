import difflib

a = ["one", "two", "three", "four"]
b = ["one", "tow", "three", "five", "six"]
print("\n".join(difflib.ndiff(a, b)))
print("\n".join(difflib.unified_diff(a, b, "a.txt", "b.txt", lineterm="")))
print("\n".join(difflib.context_diff(a, b, "a.txt", "b.txt", lineterm="")))
sm = difflib.SequenceMatcher(None, "abxcd", "abcd")
print(sm.ratio(), sm.get_matching_blocks())
print(sm.get_opcodes())
print(difflib.get_close_matches("appel", ["ape", "apple", "peach", "puppy"]))
print(list(difflib.Differ().compare(["a\n", "b\n"], ["a\n", "c\n"])))
print(difflib.SequenceMatcher(None, "private", "privet").quick_ratio())

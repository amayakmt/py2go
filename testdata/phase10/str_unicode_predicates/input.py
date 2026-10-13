chars = ["a", "Z", "5", "²", "½", "٣", "१", "Ⅷ", "①", " ", "\t", " ", " ", "​", "\n", "_", "é", "日", "ǅ", "́", "€", "\x00", "\x7f", "\x85", " ", "퟿", "😀", "\U0001d7ce", "ª", "ʰ", "ℵ", "ⅎ", "﻿"]
for c in chars:
    print(repr(c), c.isalpha(), c.isdigit(), c.isdecimal(), c.isnumeric(), c.isalnum(), c.isspace(), c.isprintable(), c.isidentifier(), c.isascii(), c.isupper(), c.islower(), c.istitle())
for s in ["abc", "_x1", "1x", "", "x y", "é1", "日本", "class", "𝐱", "x́", "a-b", "__init__", "٣a", "ª"]:
    print(repr(s), s.isidentifier(), s.isprintable(), s.isalnum(), s.isdecimal())
print(int("٣٤"), int("१२३"), float("١.٥"), int("²") if False else 0, "٣".isdigit(), "²".isdecimal(), "²".isdigit(), "²".isnumeric())
for bad in ("²", "½", "Ⅷ"):
    try:
        int(bad)
    except ValueError as e:
        print(e)

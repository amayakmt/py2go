s: str = "banana"
print(s.find("an"), s.find("z"), s.find(""), s.find("a"))
print(s.replace("a", "o"), s.replace("an", ""), s.replace("x", "y"), s.replace("", "-"))
print(s.count("a"), s.count("an"), s.count("z"), s.count(""))
print(s.startswith("ban"), s.startswith("nan"), s.endswith("ana"), s.endswith("ban"))
print("aaaa".count("aa"), "aaaa".replace("aa", "b"))

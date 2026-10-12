s: str = "  Hello, World  "
print(s.upper())
print(s.lower())
print("[" + s.strip() + "]")
print("xxhixx".strip("x"))
print("[" + "\t\n x \r\n".strip() + "]")
print("abc".strip("cba") == "")
print("MiXeD".lower().upper())

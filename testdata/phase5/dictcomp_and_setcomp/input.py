print({x: x * x for x in range(5)})
print({x: x * x for x in range(10) if x % 3 == 0})
print({s: len(s) for s in ["a", "bb", "ccc"]})
d: dict[str, int] = {"a": 1, "b": 2, "c": 3}
print({v: k for k, v in d.items()})
print({k: v * 10 for k, v in d.items() if v != 2})
print({(x, y): x * y for x in range(2) for y in range(2)})
print({x % 3 for x in range(10)})
print({c for c in "hello world" if c != " "} == set("helo wrd") - {" "})
print(sorted({len(w) for w in ["aa", "b", "cc", "ddd"]}))
print({x: [y for y in range(x)] for x in range(4)})
words: list[str] = ["apple", "avocado", "banana", "blueberry", "cherry"]
by_letter: dict[str, list[str]] = {}
for w in words:
    by_letter.setdefault(w[0], []).append(w)
print({k: len(v) for k, v in by_letter.items()})

import pprint

data = {"name": "x" * 20, "items": list(range(30)), "nested": {"a": [1, 2, {"b": (3, 4)}], "c": "text" * 10}}
pprint.pprint(data)
pprint.pprint(data, width=40, indent=2)
print(pprint.pformat([1, 2, 3]))
print(pprint.pformat("a" * 100, width=30))
print(pprint.isreadable([1, 2]), pprint.isrecursive([1]))
print(pprint.pformat({"b": 1, "a": 2}, sort_dicts=False))
print(pprint.pformat(list(range(40)), compact=True, width=50))

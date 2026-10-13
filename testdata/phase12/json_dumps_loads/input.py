import json
doc = {"name": "Zoë", "tags": ["a", "b"], "n": 1, "f": 2.5, "t": True, "nil": None, "nested": {"k": [1, {"z": []}]}, "big": 2 ** 70, "neg": -3, "e": 1e22, "s": "line\nbreak\ttab \"q\" \\ / \x01 \u2028 日本 \U0001F600"}
print(json.dumps(doc))
print(json.dumps(doc, ensure_ascii=False))
print(json.dumps(doc, sort_keys=True, indent=2))
print(json.dumps(doc, separators=(",", ":"), sort_keys=True))
print(json.dumps([1, 2, [3, [4]]], indent=4), json.dumps([], indent=2), json.dumps({}, indent=2), json.dumps([[]], indent=1), json.dumps({"a": {}}, indent="\t"))
print(json.dumps(1), json.dumps("s"), json.dumps(None), json.dumps(True), json.dumps(1.5), json.dumps(1.0), json.dumps(-0.0), json.dumps(10 ** 30), json.dumps((1, 2)), json.dumps({1: 2, "a": 3}), json.dumps({True: 1, None: 2, 1.5: 3}))
print(json.loads(json.dumps(doc)) == doc, json.loads('{"b": 1, "a": 2}'), list(json.loads('{"b": 1, "a": 2}')), json.loads("[1, 2.5, -3, 1e2, 1E-2, 0.5e1, -0, 12345678901234567890]"), json.loads('"\\u0041\\n\\\\\\"\\/"'), json.loads("  [ ]  "), json.loads("null"), json.loads("1"), json.loads('"\\ud83d\\ude00"'))
print(json.loads(b'{"a": "\xc3\xa9"}'), json.loads(bytearray(b"[1]")), json.loads("[NaN, Infinity, -Infinity]"), json.loads('{"a": {"b": {"c": [true, false, null]}}}'))
print(json.loads("[1, 2]", parse_int=float), json.loads('{"x": 1.5}', parse_float=lambda s: s + "f"), json.loads("[NaN]", parse_constant=lambda c: "const:" + c), json.loads('{"a": 1, "b": 2}', object_hook=lambda d: sum(d.values())), json.loads('{"a": 1, "b": 2}', object_pairs_hook=lambda p: p))

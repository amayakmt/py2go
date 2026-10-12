def render(value, indent=0):
    pad = "  " * indent
    if value is None:
        return "null"
    if value is True:
        return "true"
    if value is False:
        return "false"
    if isinstance(value, (int, float)):
        return str(value)
    if isinstance(value, str):
        return '"' + value.replace('"', '\\"') + '"'
    if isinstance(value, (list, tuple)):
        if not value:
            return "[]"
        inner = ",\n".join(pad + "  " + render(v, indent + 1) for v in value)
        return "[\n" + inner + "\n" + pad + "]"
    if isinstance(value, dict):
        if not value:
            return "{}"
        inner = ",\n".join(pad + "  " + render(str(k), 0) + ": " + render(v, indent + 1) for k, v in sorted(value.items(), key=lambda kv: str(kv[0])))
        return "{\n" + inner + "\n" + pad + "}"
    raise TypeError("cannot render " + type(value).__name__)

doc = {"name": "x", "tags": ["a", "b"], "n": 3, "pi": 3.14, "ok": True, "none": None,
       "nested": {"empty_list": [], "empty_dict": {}, "t": (1, 2)}, 5: "five", "q": 'say "hi"'}
print(render(doc))
print(render([]), render("s"), render([[1], [2, [3]]]))
try:
    render({1, 2})
except TypeError as e:
    print(e)

def depth(v):
    if isinstance(v, dict):
        return 1 + max([depth(x) for x in v.values()] or [0])
    if isinstance(v, (list, tuple)):
        return 1 + max([depth(x) for x in v] or [0])
    return 0
print(depth(doc), depth([]), depth(1), depth([[[]]]))

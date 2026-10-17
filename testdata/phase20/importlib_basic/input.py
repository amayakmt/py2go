import importlib
import importlib.util
import json
import textwrap  # linked, so that import_module(name) finds it

m = importlib.import_module("json")
print(m is json, m.dumps([1, 2]))
cs = importlib.import_module("colorsys")
print(cs.rgb_to_hsv(1.0, 0.0, 0.0))
name = "textwrap"
tw = importlib.import_module(name)
print(tw.shorten("hello world and more", width=12))


def load(n):
    return importlib.import_module(n)


print(load("json").loads("[3]"), load("json") is json)
try:
    importlib.import_module("no_such_module_xyz")
except ModuleNotFoundError as e:
    print("ModuleNotFoundError", e, e.name)
print(importlib.util.find_spec("json") is not None, importlib.util.find_spec("no_such_module_xyz"))
print(importlib.util.find_spec("json").name, importlib.reload(json) is json)
print(importlib.util.resolve_name(".decoder", "json"), importlib.import_module(".decoder", "json").JSONDecoder.__name__)
try:
    importlib.import_module(".x")
except TypeError as e:
    print("TypeError", e)

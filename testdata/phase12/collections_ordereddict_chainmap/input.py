from collections import OrderedDict, ChainMap
o = OrderedDict([("a", 1), ("b", 2)])
o["c"] = 3; o["a"] = 10
print(o, list(o), list(o.items()), len(o), o["a"], "b" in o, o.get("z", 0), OrderedDict(), OrderedDict(x=1))
o.move_to_end("a"); print(list(o))
o.move_to_end("c", last=False); print(list(o))
print(o.popitem(), o.popitem(last=False), list(o), o == OrderedDict([("b", 2)]), OrderedDict([("a", 1), ("b", 2)]) == OrderedDict([("b", 2), ("a", 1)]), OrderedDict(a=1, b=2) == {"b": 2, "a": 1})
del o["b"]
print(o, len(o))
try:
    o.popitem()
except KeyError as e:
    print("KeyError", e)
try:
    o["nope"]
except KeyError as e:
    print("KeyError", e)
defaults = {"color": "red", "user": "guest"}
cli = {"color": "blue"}
cm = ChainMap(cli, defaults)
print(cm["color"], cm["user"], cm.get("x", "none"), "user" in cm, "q" in cm, len(cm), sorted(cm), sorted(cm.items()))
cm["color"] = "green"; cm["new"] = 1
print(cli, defaults, cm.maps)
child = cm.new_child({"color": "child"})
print(child["color"], child["user"], len(child.maps), child.parents["color"])
del cm["new"]
print(cli)
try:
    cm["missing"]
except KeyError as e:
    print("KeyError", e)

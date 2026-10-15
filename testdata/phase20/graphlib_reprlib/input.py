import graphlib
import reprlib

ts = graphlib.TopologicalSorter({"d": {"b", "c"}, "c": {"a"}, "b": {"a"}})
print(list(ts.static_order()))
ts2 = graphlib.TopologicalSorter()
ts2.add("x", "y", "z")
ts2.add("y", "z")
ts2.prepare()
order = []
while ts2.is_active():
    ready = ts2.get_ready()
    order.append(sorted(ready))
    ts2.done(*ready)
print(order)
try:
    graphlib.TopologicalSorter({"a": {"b"}, "b": {"a"}}).prepare()
except graphlib.CycleError as e:
    print("CycleError", e.args[0])

r = reprlib.Repr()
r.maxlist = 3
print(r.repr([1, 2, 3, 4, 5]))
print(reprlib.repr(list(range(100))))
print(reprlib.repr("x" * 100))
print(reprlib.repr({"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}))
print(reprlib.repr(set(range(20))))
print(reprlib.repr((1, 2, 3, 4, 5, 6, 7, 8)))
print(reprlib.repr(12345678901234567890123456789012345678901234567890))

import json
for bad in ("", " ", "{", "}", '{"a"}', '{"a":}', '{"a" 1}', "[1,]", "[,1]", "[1 2]", '{"a": 1,}', "tru", "nul", "'x'", '"abc', '"a\\x"', '"a\\u12"', "1 2", "[1]]", '{1: 2}', "\n\n  x", "-", "01", "1.", ".5", "+1", "[1.e3]", '"tab\there"', "\ufeff[]", "NaN1", "[", '{"a": [1, 2}', "00"):
    try:
        print(repr(bad), "->", json.loads(bad))
    except json.JSONDecodeError as e:
        print(repr(bad), "|", e.msg, "|", e.pos, e.lineno, e.colno, "|", str(e), "|", e.doc == bad, isinstance(e, ValueError))
try:
    json.loads(5)
except TypeError as e:
    print(e)
try:
    json.loads(None)
except TypeError as e:
    print(e)
class Thing:
    def __init__(self, v):
        self.v = v
def encode_thing(o):
    if isinstance(o, Thing):
        return {"thing": o.v}
    raise TypeError("cannot encode " + type(o).__name__)
print(json.dumps([Thing(1), {"k": Thing([2])}], default=encode_thing), json.dumps({"a": {1, 2}} if False else {"a": 1}, default=str))
for obj in (Thing(1), {Thing(1): 1}, {1, 2}, [object()], b"x", float("nan"), float("-inf")):
    try:
        print(json.dumps(obj))
    except (TypeError, ValueError) as e:
        print(type(e).__name__, e)
for kw in ({"allow_nan": False}, {"skipkeys": True}):
    try:
        print(json.dumps([float("nan"), {(1,): 2}], **kw))
    except (TypeError, ValueError) as e:
        print(type(e).__name__, e)
loop = {}
loop["me"] = loop
try:
    json.dumps(loop)
except ValueError as e:
    print(e)
class Enc(json.JSONEncoder):
    def default(self, o):
        return "custom:" + type(o).__name__
print(json.dumps({"x": Thing(1)}, cls=Enc), json.JSONEncoder().encode([1, "a"]), json.JSONDecoder().decode('{"q": [1]}'))
import io
buf = io.StringIO()
json.dump({"a": [1, 2]}, buf, indent=1)
print(repr(buf.getvalue()), json.load(io.StringIO('{"z": 1}')))

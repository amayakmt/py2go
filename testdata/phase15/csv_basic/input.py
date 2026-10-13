import csv
import io

def read(text, **kw):
    return list(csv.reader(io.StringIO(text, newline=""), **kw))

print(read("a,b,c\r\n1,2,3\r\n"), read("a,\"b,c\",d\n"), read('x,"he said ""hi""",y\n'), read("a,b\n\n\nc,d\n"), read(""), read("\n"), read("a,b"), read("a;b;c\n", delimiter=";"))
print(read('"multi\nline",x\n'), read("a, b, c\n", skipinitialspace=True), read("a,b\\,c\n", escapechar="\\"), read("'a,b',c\n", quotechar="'"), read('a,"b"x\n'), read('1,2.5,"x"\n', quoting=csv.QUOTE_NONNUMERIC), read(",,\n"), read('a\tb\n', dialect="excel-tab"))
print(read('"a",b\n', quoting=csv.QUOTE_NOTNULL), read(',"",x\n', quoting=csv.QUOTE_NOTNULL), read(',"",1\n', quoting=csv.QUOTE_STRINGS), read('a,"b""c"\n', doublequote=False, escapechar="\\") if False else "")
for text, kw in [('a,"b\n', {"strict": True}), ('"a"b,c\n', {"strict": True}), ("a\rb\n", {}), ('1,x\n', {"quoting": csv.QUOTE_NONNUMERIC}), ("a\0b\n", {}), ("a,b\n", {"delimiter": "ab"}), ("a,b\n", {"quoting": 9}), ("a", {"dialect": "nope"}), ("a", {"bogus": 1})]:
    try:
        print(read(text, **kw))
    except (csv.Error, TypeError, ValueError) as e:
        print(type(e).__name__, e)
r = csv.reader(["a,b", "c,d"])
print(next(r), r.line_num, next(r), r.line_num, r.dialect.delimiter)

def write(rows, **kw):
    out = io.StringIO()
    w = csv.writer(out, **kw)
    w.writerows(rows)
    return out.getvalue()

print(repr(write([["a", "b c", "d,e", 'f"g', "", None, 1, 2.5, True, "x\ny"]])), repr(write([[""]])), repr(write([[]])), repr(write([["a"]], lineterminator="\n")), repr(write([["a", 1, 2.0, None]], quoting=csv.QUOTE_ALL)), repr(write([["a", 1, 2.0, None]], quoting=csv.QUOTE_NONNUMERIC)))
print(repr(write([["a b", "c,d", 'e"f']], quoting=csv.QUOTE_NONE, escapechar="\\")), repr(write([['a"b']], doublequote=False, escapechar="\\")), repr(write([["a", "b"]], delimiter=";", quotechar="'")), repr(write([["a'b;c"]], delimiter=";", quotechar="'")), repr(write([["x", None, "y"]], quoting=csv.QUOTE_STRINGS)), repr(write([["x", None, ""]], quoting=csv.QUOTE_NOTNULL)), repr(write([["a"]], dialect="unix")), repr(write([["a", "b"]], dialect="excel-tab")))
for rows, kw in [([["a,b"]], {"quoting": csv.QUOTE_NONE}), ([[""]], {"quoting": csv.QUOTE_NONE}), ([5], {}), ([["a"]], {"lineterminator": 5})]:
    try:
        print(repr(write(rows, **kw)))
    except (csv.Error, TypeError) as e:
        print(type(e).__name__, e)
class Stub:
    def write(self, s):
        return len(s)
print(csv.writer(Stub()).writerow(["ab", "c"]), csv.field_size_limit(), csv.list_dialects(), csv.QUOTE_MINIMAL, csv.QUOTE_NONE, csv.__version__)
csv.register_dialect("pipes", delimiter="|", quoting=csv.QUOTE_MINIMAL)
print(read("a|b|\"c|d\"\n", dialect="pipes"), "pipes" in csv.list_dialects(), csv.get_dialect("pipes").delimiter)
csv.unregister_dialect("pipes")
class MyD(csv.Dialect):
    delimiter = "#"
    quotechar = '"'
    doublequote = True
    skipinitialspace = False
    lineterminator = "\n"
    quoting = csv.QUOTE_MINIMAL
print(read("a#b#c\n", dialect=MyD), repr(write([["a", "b#c"]], dialect=MyD)))
dr = csv.DictReader(io.StringIO("name,age\nann,30\nbob,25,extra\ncy\n\ndee,1\n", newline=""))
print(dr.fieldnames, list(dr), dr.line_num)
print(list(csv.DictReader(io.StringIO("1,2\n3,4\n"), fieldnames=["x", "y"])), list(csv.DictReader(io.StringIO("a,b\n1\n"), restval="?")), list(csv.DictReader(io.StringIO("a\n1,2\n"), restkey="rest")))
out = io.StringIO()
dw = csv.DictWriter(out, fieldnames=["id", "name"], restval="-")
dw.writeheader()
dw.writerow({"id": 1, "name": "x"})
dw.writerows([{"id": 2}, {"name": "z"}])
print(repr(out.getvalue()))
try:
    dw.writerow({"id": 1, "bad": 2})
except ValueError as e:
    print("ValueError", e)
print(csv.DictWriter(io.StringIO(), ["a"], extrasaction="ignore").writerow({"a": 1, "zz": 2}))
with open("/tmp/py2go_csv_test.csv", "w", newline="") as f:
    csv.writer(f).writerows([["h1", "h2"], ["v 1", "v,2"]])
with open("/tmp/py2go_csv_test.csv", newline="") as f:
    print(list(csv.reader(f)))
